// operations.go
package backup

import (
	"context"
	"fmt"
	"os"
	"time"
)

func (s *Service) Backup(ctx context.Context) error {
	// Create backup tasks
	tasks, totalFiles, err := s.createTasks()
	if err != nil {
		return err
	}

	if !s.config.Options.Quiet {
		fmt.Printf("Starting backup of %d files...\n", totalFiles)
	}

	// Initialize metrics and start tracking
	s.metrics = NewBackupMetrics(totalFiles, s.config.Options.Quiet)
	s.metrics.StartTracking(ctx)

	// Start new backup version
	s.versioner.StartNewVersion(s.config)

	// Create a done channel for the display goroutine
	done := make(chan struct{})
	defer close(done)

	// Start progress display in a separate goroutine
	if !s.config.Options.Quiet {
		go func() {
			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					s.metrics.DisplayProgress()
				case <-done:
					s.metrics.DisplayProgress() // One final update
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Execute backup
	err = s.pool.Execute(ctx, tasks)

	// Wait a moment for final progress update
	time.Sleep(200 * time.Millisecond)

	// Get final stats and complete version
	stats := s.metrics.GetStats()
	if err := s.versioner.CompleteVersion(stats); err != nil {
		s.logger.Error("Failed to save backup version: %v", err)
	}

	// Print final summary
	s.metrics.DisplayFinalSummary()

	// Close the metrics updates channel
	close(s.metrics.updates)

	return err
}

// DryRun simulates the backup process without making changes
func (s *Service) DryRun(ctx context.Context) error {
	// Validate paths
	if err := s.validatePaths(); err != nil {
		return err
	}

	// Create backup tasks
	tasks, totalFiles, err := s.createTasks()
	if err != nil {
		return err
	}

	// Initialize new metrics for this dry run
	s.metrics = NewBackupMetrics(totalFiles, s.config.Options.Quiet)

	totalSize := int64(0)
	fileCount := 0
	skippedCount := 0
	skippedSize := int64(0)

	// Simulate the backup
	for _, task := range tasks {
		if skip, err := s.shouldSkipFile(task); err != nil {
			s.logger.Warn("Cannot check file %s: %v", task.Source, err)
			continue
		} else if skip {
			skippedCount++
			info, _ := os.Stat(task.Source)
			skippedSize += info.Size()
			if !s.config.Options.Quiet {
				fmt.Printf("Would skip: %s (identical)\n", task.Source)
			}
			continue
		}

		info, err := os.Stat(task.Source)
		if err != nil {
			s.logger.Warn("Cannot stat file %s: %v", task.Source, err)
			continue
		}

		totalSize += info.Size()
		fileCount++

		if !s.config.Options.Quiet {
			fmt.Printf("Would copy: %s -> %s\n", task.Source, task.Destination)
		}
	}

	duration := s.metrics.GetDuration()
	if !s.config.Options.Quiet {
		fmt.Printf("\nDry run completed in %v\n", duration)
		fmt.Printf("Files to copy: %d, Files to skip: %d\n", fileCount, skippedCount)
		fmt.Printf("Data to copy: %.2f MB, Data to skip: %.2f MB\n",
			float64(totalSize)/1024/1024,
			float64(skippedSize)/1024/1024)
	}

	return nil
}
