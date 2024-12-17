package backup

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Service represents the backup service with all required dependencies
type Service struct {
	config  *Config
	logger  *Logger
	metrics *Metrics
	pool    *WorkerPool
}

// CopyTask represents a single file copy operation
type CopyTask struct {
	Source      string
	Destination string
	Size        int64
	ModTime     time.Time
}

// Metrics tracks backup operation statistics
type Metrics struct {
	mu           sync.Mutex
	BytesCopied  int64
	FilesCopied  int
	FilesSkipped int
	Errors       int
	StartTime    time.Time
	EndTime      time.Time
}

// FileInfo holds metadata about a file being backed up
type FileInfo struct {
	Path        string
	Size        int64
	ModTime     time.Time
	Checksum    string
	IsDirectory bool
}

// NewService creates a new backup service instance
func NewService(cfg *Config) (*Service, error) {
	logger, err := NewLogger(cfg.TargetDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %v", err)
	}

	s := &Service{
		config:  cfg,
		logger:  logger,
		metrics: &Metrics{},
	}

	s.pool = NewWorkerPool(cfg.Concurrency, s.copyFile)
	return s, nil
}

// Backup performs the backup operation
func (s *Service) Backup(ctx context.Context) error {
	// Store the start time before any operations
	startTime := time.Now()

	// Initialize metrics
	s.metrics.mu.Lock()
	s.metrics.StartTime = startTime
	s.metrics.BytesCopied = 0
	s.metrics.FilesCopied = 0
	s.metrics.FilesSkipped = 0
	s.metrics.Errors = 0
	s.metrics.mu.Unlock()

	// Validate paths
	if err := s.validatePaths(); err != nil {
		return err
	}

	// Create backup tasks first
	tasks, err := s.createTasks()
	if err != nil {
		return err
	}

	if !s.config.Options.Quiet {
		fmt.Printf("Starting backup of %d files...\n", len(tasks))
	}

	// Execute backup
	if err := s.pool.Execute(ctx, tasks); err != nil {
		return err
	}

	// Calculate duration properly
	duration := time.Since(startTime)

	// Update final metrics
	s.metrics.mu.Lock()
	s.metrics.EndTime = time.Now()
	totalBytes := s.metrics.BytesCopied
	filesCopied := s.metrics.FilesCopied
	filesSkipped := s.metrics.FilesSkipped
	errors := s.metrics.Errors
	s.metrics.mu.Unlock()

	// Log completion with proper formatting
	s.logger.Info("Backup completed in %v. Files copied: %d, Files skipped: %d, Errors: %d, Total size: %.2f MB",
		duration,
		filesCopied,
		filesSkipped,
		errors,
		float64(totalBytes)/1024/1024)

	// Print user-friendly summary
	if !s.config.Options.Quiet {
		fmt.Printf("\nBackup completed successfully in %v\n", duration)
		fmt.Printf("Files copied: %d, Files skipped: %d, Total size: %.2f MB\n",
			filesCopied,
			filesSkipped,
			float64(totalBytes)/1024/1024)
	}

	return nil
}

// validatePaths ensures all necessary directories exist
func (s *Service) validatePaths() error {
	// Check source directory
	if _, err := os.Stat(s.config.SourceDirectory); err != nil {
		return newBackupError("ValidateSource", s.config.SourceDirectory, err)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(s.config.TargetDirectory, 0755); err != nil {
		return newBackupError("CreateTarget", s.config.TargetDirectory, err)
	}

	return nil
}

// createTasks generates the list of files to be backed up
func (s *Service) createTasks() ([]CopyTask, error) {
	var tasks []CopyTask

	for _, folder := range s.config.FoldersToBackup {
		srcPath := filepath.Join(s.config.SourceDirectory, folder)
		dstPath := filepath.Join(s.config.TargetDirectory, folder)

		err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip if matches exclude patterns
			for _, pattern := range s.config.ExcludePatterns {
				if matched, _ := filepath.Match(pattern, info.Name()); matched {
					s.logger.Debug("Skipping excluded file: %s", path)
					return nil
				}
			}

			// Create relative path
			relPath, err := filepath.Rel(srcPath, path)
			if err != nil {
				return err
			}

			destPath := filepath.Join(dstPath, relPath)

			if !info.IsDir() {
				tasks = append(tasks, CopyTask{
					Source:      path,
					Destination: destPath,
					Size:        info.Size(),
					ModTime:     info.ModTime(),
				})
			}

			return nil
		})

		if err != nil {
			return nil, newBackupError("CreateTasks", srcPath, err)
		}
	}

	return tasks, nil
}

// copyFile performs the actual file copy operation
func (s *Service) copyFile(task CopyTask) error {
	// Skip if destination exists and is identical
	if s.config.DeepDuplicateCheck {
		identical, err := s.compareFiles(task.Source, task.Destination)
		if err != nil {
			s.logger.Warn("Failed to compare files %s and %s: %v", task.Source, task.Destination, err)
		} else if identical {
			s.metrics.mu.Lock()
			s.metrics.FilesSkipped++
			s.metrics.mu.Unlock()
			s.logger.Debug("Skipped identical file: %s", task.Source)
			return nil
		}
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(task.Destination)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return newBackupError("CreateDir", destDir, err)
	}

	// Attempt copy with retries
	var lastErr error
	for attempt := 0; attempt < s.config.RetryAttempts; attempt++ {
		if err := s.performCopy(task); err != nil {
			lastErr = err
			s.logger.Warn("Copy attempt %d failed for %s: %v", attempt+1, task.Source, err)
			time.Sleep(s.config.RetryDelay)
			continue
		}
		return nil
	}

	s.metrics.mu.Lock()
	s.metrics.Errors++
	s.metrics.mu.Unlock()

	return newBackupError("CopyFile", task.Source, lastErr)
}

// performCopy executes a single copy operation
func (s *Service) performCopy(task CopyTask) error {
	src, err := os.Open(task.Source)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(task.Destination)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy with progress tracking
	buf := make([]byte, s.config.BufferSize)
	hasher := sha256.New()
	writer := io.MultiWriter(dst, hasher)

	copied, err := io.CopyBuffer(writer, src, buf)
	if err != nil {
		return err
	}

	// Update metrics
	s.metrics.mu.Lock()
	s.metrics.BytesCopied += copied
	s.metrics.FilesCopied++
	s.metrics.mu.Unlock()

	// Preserve file mode
	if sourceInfo, err := os.Stat(task.Source); err == nil {
		if err := os.Chmod(task.Destination, sourceInfo.Mode()); err != nil {
			s.logger.Warn("Failed to preserve file mode for %s: %v", task.Destination, err)
		}
	}

	// Log with current timestamp
	s.logger.Info("Copied %s (%.2f MB) at %s", task.Source, float64(copied)/1024/1024, time.Now().Format("15:04:05.000"))

	return nil
}

// compareFiles checks if two files are identical
func (s *Service) compareFiles(src, dst string) (bool, error) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false, err
	}

	dstInfo, err := os.Stat(dst)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if srcInfo.Size() != dstInfo.Size() {
		return false, nil
	}

	if s.config.DeepDuplicateCheck {
		srcHash, err := s.calculateChecksum(src)
		if err != nil {
			return false, err
		}

		dstHash, err := s.calculateChecksum(dst)
		if err != nil {
			return false, err
		}

		return srcHash == dstHash, nil
	}

	return srcInfo.ModTime() == dstInfo.ModTime(), nil
}

// calculateChecksum computes the SHA-256 hash of a file
func (s *Service) calculateChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// DryRun simulates the backup process without making changes
func (s *Service) DryRun(ctx context.Context) error {
	s.metrics.StartTime = time.Now()
	defer func() { s.metrics.EndTime = time.Now() }()

	// Validate paths
	if err := s.validatePaths(); err != nil {
		return err
	}

	// Create backup tasks
	tasks, err := s.createTasks()
	if err != nil {
		return err
	}

	totalSize := int64(0)
	fileCount := 0

	// Simulate the backup
	for _, task := range tasks {
		info, err := os.Stat(task.Source)
		if err != nil {
			s.logger.Warn("Cannot stat file %s: %v", task.Source, err)
			continue
		}

		totalSize += info.Size()
		fileCount++

		s.logger.Info("[DRY RUN] Would copy: %s -> %s (%.2f MB)",
			task.Source, task.Destination, float64(info.Size())/1024/1024)

		if !s.config.Options.Quiet {
			fmt.Printf("Would copy: %s -> %s\n", task.Source, task.Destination)
		}
	}

	duration := time.Since(s.metrics.StartTime)
	s.logger.Info("Dry run completed. Would copy %d files, total size: %.2f MB",
		fileCount, float64(totalSize)/1024/1024)

	if !s.config.Options.Quiet {
		fmt.Printf("\nDry run completed in %v\n", duration)
		fmt.Printf("Files to copy: %d, Total size: %.2f MB\n",
			fileCount, float64(totalSize)/1024/1024)
	}
	return nil
}
