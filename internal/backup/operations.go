// operations.go
package backup

import (
	"context"
	"fmt"
	"os"
	"time"
)

func (s *Service) Backup(ctx context.Context) error {
	startTime := time.Now()

	// Start new backup version
	version := s.versioner.StartNewVersion(s.config)

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
		version.Status = "Failed"
		return err
	}

	// Create backup tasks
	tasks, err := s.createTasks()
	if err != nil {
		version.Status = "Failed"
		return err
	}

	if !s.config.Options.Quiet {
		fmt.Printf("Starting backup of %d files...\n", len(tasks))
	}

	// Execute backup
	if err := s.pool.Execute(ctx, tasks); err != nil {
		version.Status = "Failed"
		s.versioner.CompleteVersion(BackupStats{
			TotalFiles:       len(tasks),
			FilesBackedUp:    s.metrics.FilesCopied,
			FilesSkipped:     s.metrics.FilesSkipped,
			FilesFailed:      s.metrics.Errors,
			TotalBytes:       s.metrics.BytesCopied,
			BytesTransferred: s.metrics.BytesCopied,
		})
		return err
	}

	// Calculate duration properly
	duration := time.Since(startTime)

	// Update final metrics and complete version
	s.metrics.mu.Lock()
	stats := BackupStats{
		TotalFiles:       len(tasks),
		FilesBackedUp:    s.metrics.FilesCopied,
		FilesSkipped:     s.metrics.FilesSkipped,
		FilesFailed:      s.metrics.Errors,
		TotalBytes:       s.metrics.BytesCopied,
		BytesTransferred: s.metrics.BytesCopied,
	}
	s.metrics.mu.Unlock()

	if err := s.versioner.CompleteVersion(stats); err != nil {
		s.logger.Error("Failed to save backup version: %v", err)
	}

	// Log completion
	s.logger.Info("Backup completed in %v. Files copied: %d, Files skipped: %d, Errors: %d, Total size: %.2f MB",
		duration,
		stats.FilesBackedUp,
		stats.FilesSkipped,
		stats.FilesFailed,
		float64(stats.TotalBytes)/1024/1024)

	// Print user-friendly summary
	if !s.config.Options.Quiet {
		fmt.Printf("\nBackup completed successfully in %v\n", duration)
		fmt.Printf("Files copied: %d, Files skipped: %d, Total size: %.2f MB\n",
			stats.FilesBackedUp,
			stats.FilesSkipped,
			float64(stats.TotalBytes)/1024/1024)
	}
	return nil
}

// DryRun simulates the backup process without making changes
func (s *Service) DryRun(ctx context.Context) error {
	s.metrics.StartTime = time.Now()

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
