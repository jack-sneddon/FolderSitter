// operations.go
package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	// Validate only source path exists
	if _, err := os.Stat(s.config.SourceDirectory); err != nil {
		return fmt.Errorf("source directory does not exist: %v", err)
	}

	// Create backup tasks
	tasks, totalFiles, err := s.createTasks()
	if err != nil {
		return err
	}

	// Create log file in system temp directory
	logFile := filepath.Join(os.TempDir(),
		fmt.Sprintf("foldersitter_dryrun_%s.log",
			time.Now().Format("2006-01-02_15-04-05")))

	// Initialize metrics and counters
	s.metrics = NewBackupMetrics(totalFiles, s.config.Options.Quiet)
	s.metrics.StartTracking(ctx)
	totalSize := int64(0)
	fileCount := 0
	skippedCount := 0
	skippedSize := int64(0)

	// Create a done channel for the display goroutine
	done := make(chan struct{})
	defer close(done)

	// Start progress display
	if !s.config.Options.Quiet {
		fmt.Printf("Starting dry run analysis of %d files...\n\n", totalFiles)
		go func() {
			ticker := time.NewTicker(200 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					displayDryRunProgress(totalFiles, fileCount+skippedCount)
				case <-done:
					displayDryRunProgress(totalFiles, fileCount+skippedCount)
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// Open log file for writing
	file, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}
	defer file.Close()

	// Write log header
	fmt.Fprintf(file, "FolderSitter Dry Run Analysis\n")
	fmt.Fprintf(file, "Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(file, "Source: %s\n", s.config.SourceDirectory)
	fmt.Fprintf(file, "Target: %s\n", s.config.TargetDirectory)
	fmt.Fprintf(file, "----------------------------------------\n\n")

	// Log details and collect statistics
	for _, task := range tasks {
		if _, err := os.Stat(s.config.TargetDirectory); os.IsNotExist(err) {
			// Target doesn't exist, all files need to be copied
			info, err := os.Stat(task.Source)
			if err != nil {
				fmt.Fprintf(file, "ERROR: Cannot stat file %s: %v\n", task.Source, err)
				continue
			}
			totalSize += info.Size()
			fileCount++
			fmt.Fprintf(file, "COPY: %s -> %s (%.2f MB)\n",
				task.Source, task.Destination, float64(info.Size())/1024/1024)
		} else {
			// Target exists, check for identical files
			if skip, err := s.shouldSkipFile(task); err != nil {
				fmt.Fprintf(file, "ERROR: Cannot check file %s: %v\n", task.Source, err)
				continue
			} else if skip {
				skippedCount++
				info, _ := os.Stat(task.Source)
				skippedSize += info.Size()
				fmt.Fprintf(file, "SKIP: %s (identical)\n", task.Source)
				continue
			}

			info, err := os.Stat(task.Source)
			if err != nil {
				fmt.Fprintf(file, "ERROR: Cannot stat file %s: %v\n", task.Source, err)
				continue
			}

			totalSize += info.Size()
			fileCount++
			fmt.Fprintf(file, "COPY: %s -> %s (%.2f MB)\n",
				task.Source, task.Destination, float64(info.Size())/1024/1024)
		}
	}

	// Write summary to log
	fmt.Fprintf(file, "\n----------------------------------------\n")
	fmt.Fprintf(file, "Summary:\n")
	fmt.Fprintf(file, "Files to copy: %d (%.2f MB)\n", fileCount, float64(totalSize)/1024/1024)
	fmt.Fprintf(file, "Files to skip: %d (%.2f MB)\n", skippedCount, float64(skippedSize)/1024/1024)

	// Allow progress bar to complete
	time.Sleep(200 * time.Millisecond)

	// Display console summary
	if !s.config.Options.Quiet {
		duration := s.metrics.GetDuration()
		fmt.Printf("\n\nDry run completed in %v\n", duration)
		fmt.Printf("Summary:\n")
		fmt.Printf("- Files to copy: %d (%.2f MB)\n", fileCount, float64(totalSize)/1024/1024)
		fmt.Printf("- Files to skip: %d (%.2f MB)\n", skippedCount, float64(skippedSize)/1024/1024)
		fmt.Printf("\nDetailed analysis has been written to:\n%s\n", logFile)
	}

	return nil
}

// Helper function for dry run progress display
func displayDryRunProgress(total, current int) {
	percentComplete := float64(current) / float64(total) * 100

	// Create progress bar
	const barWidth = 30
	completed := int(percentComplete * float64(barWidth) / 100)
	if completed < 0 {
		completed = 0
	}
	if completed > barWidth {
		completed = barWidth
	}

	bar := strings.Repeat("█", completed) + strings.Repeat("░", barWidth-completed)

	// Save cursor position, clear line, write progress
	fmt.Print("\x1b[s")     // Save cursor position
	fmt.Print("\x1b[1000D") // Move cursor far left
	fmt.Print("\x1b[K")     // Clear line
	fmt.Printf("[%s] %5.1f%% | %d/%d files analyzed",
		bar,
		percentComplete,
		current,
		total)
	fmt.Print("\x1b[u") // Restore cursor position
}
