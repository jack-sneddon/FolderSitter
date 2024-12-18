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

	// Create backup tasks
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
	totalBytes := s.metrics.BytesCopied
	filesCopied := s.metrics.FilesCopied
	filesSkipped := s.metrics.FilesSkipped
	errors := s.metrics.Errors
	s.metrics.mu.Unlock()

	// Log completion
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
