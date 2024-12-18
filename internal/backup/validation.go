// validation.go
package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// Configuration limits
	maxConcurrency   = 32               // Maximum number of concurrent workers
	minConcurrency   = 1                // Minimum number of concurrent workers
	maxRetryAttempts = 10               // Maximum number of retry attempts
	minRetryAttempts = 0                // Minimum number of retry attempts
	maxRetryDelay    = time.Hour        // Maximum delay between retries
	minRetryDelay    = time.Second      // Minimum delay between retries
	maxBufferSize    = 10 * 1024 * 1024 // 10MB maximum buffer size
	minBufferSize    = 4 * 1024         // 4KB minimum buffer size
)

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

// shouldSkipFile determines if a file should be skipped based on metadata and checksum
// validations.go
func (s *Service) shouldSkipFile(task CopyTask) (bool, error) {
	sourceInfo, err := os.Stat(task.Source)
	if err != nil {
		return false, fmt.Errorf("failed to stat source file: %w", err)
	}

	destInfo, err := os.Stat(task.Destination)
	if os.IsNotExist(err) {
		s.logger.Debug("Destination file does not exist: %s", task.Destination)
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to stat destination file: %w", err)
	}

	// Quick size comparison first
	if sourceInfo.Size() != destInfo.Size() {
		s.logger.Debug("Size mismatch - Source: %d bytes, Destination: %d bytes",
			sourceInfo.Size(), destInfo.Size())
		return false, nil
	}

	if s.config.DeepDuplicateCheck {
		// Calculate checksums for both files
		sourceChecksum, err := s.calculateChecksum(task.Source)
		if err != nil {
			return false, fmt.Errorf("failed to calculate source checksum: %w", err)
		}

		destChecksum, err := s.calculateChecksum(task.Destination)
		if err != nil {
			return false, fmt.Errorf("failed to calculate destination checksum: %w", err)
		}

		if sourceChecksum != destChecksum {
			s.logger.Debug("Checksum mismatch - Source: %s, Destination: %s",
				sourceChecksum, destChecksum)
			return false, nil
		}
	}

	// Files are identical - update metrics
	// s.metrics.IncrementSkipped(sourceInfo.Size())

	s.logger.Debug("Skipped identical file: %s (Size: %.2f MB)",
		task.Source, float64(sourceInfo.Size())/1024/1024)
	return true, nil
}

// validateWorkerConfig performs detailed validation of worker pool settings
func validateWorkerConfig(cfg *Config) error {
	// Validate concurrency
	if cfg.Concurrency < minConcurrency || cfg.Concurrency > maxConcurrency {
		return newBackupError(
			"ValidateWorker",
			"",
			fmt.Errorf("concurrency must be between %d and %d, got %d",
				minConcurrency, maxConcurrency, cfg.Concurrency),
		)
	}

	// Validate retry attempts
	if cfg.RetryAttempts < minRetryAttempts || cfg.RetryAttempts > maxRetryAttempts {
		return newBackupError(
			"ValidateWorker",
			"",
			fmt.Errorf("retry attempts must be between %d and %d, got %d",
				minRetryAttempts, maxRetryAttempts, cfg.RetryAttempts),
		)
	}

	// Validate retry delay
	if cfg.RetryDelay < minRetryDelay || cfg.RetryDelay > maxRetryDelay {
		return newBackupError(
			"ValidateWorker",
			"",
			fmt.Errorf("retry delay must be between %v and %v, got %v",
				minRetryDelay, maxRetryDelay, cfg.RetryDelay),
		)
	}

	// Validate buffer size
	if cfg.BufferSize < minBufferSize || cfg.BufferSize > maxBufferSize {
		return newBackupError(
			"ValidateWorker",
			"",
			fmt.Errorf("buffer size must be between %d and %d bytes, got %d",
				minBufferSize, maxBufferSize, cfg.BufferSize),
		)
	}

	return nil
}

// validateSystemResources checks if the system can handle the requested configuration
func validateSystemResources(cfg *Config) error {
	// Get number of CPU cores
	numCPU := runtime.NumCPU()

	// Ensure concurrency doesn't exceed 2x number of CPU cores
	if cfg.Concurrency > numCPU*2 {
		return newBackupError(
			"ValidateResources",
			"",
			fmt.Errorf("requested concurrency (%d) exceeds recommended maximum (%d) for %d CPU cores",
				cfg.Concurrency, numCPU*2, numCPU),
		)
	}

	// Calculate total buffer size across all workers
	totalBufferSize := int64(cfg.BufferSize) * int64(cfg.Concurrency)

	// Set a reasonable maximum total buffer size (e.g., 1GB)
	const maxTotalBufferSize = int64(1024 * 1024 * 1024) // 1GB

	if totalBufferSize > maxTotalBufferSize {
		return newBackupError(
			"ValidateResources",
			"",
			fmt.Errorf("total buffer size (%d bytes) exceeds maximum allowed (%d bytes)",
				totalBufferSize, maxTotalBufferSize),
		)
	}

	return nil
}

// Validate performs comprehensive validation of the configuration
func Validate(cfg *Config) error {
	// Basic validation
	if cfg.SourceDirectory == "" {
		return newBackupError("Validate", "", fmt.Errorf("source_directory is empty"))
	}
	if cfg.TargetDirectory == "" {
		return newBackupError("Validate", "", fmt.Errorf("target_directory is empty"))
	}
	if len(cfg.FoldersToBackup) == 0 {
		return newBackupError("Validate", "", fmt.Errorf("folders_to_backup is empty"))
	}

	// Check source directory exists
	if _, err := os.Stat(cfg.SourceDirectory); err != nil {
		return newBackupError("Validate", cfg.SourceDirectory, fmt.Errorf("source directory does not exist"))
	}

	// Worker and resource validation
	if err := validateWorkerConfig(cfg); err != nil {
		return err
	}

	if err := validateSystemResources(cfg); err != nil {
		return err
	}

	// Validate exclude patterns
	for _, pattern := range cfg.ExcludePatterns {
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return newBackupError(
				"Validate",
				pattern,
				fmt.Errorf("invalid exclude pattern: %v", err),
			)
		}
	}

	return nil
}

// ValidateConfigChange validates configuration changes at runtime
// validation.go
func (s *Service) ValidateConfigChange(newCfg *Config) error {
	// Validate the new configuration
	if err := Validate(newCfg); err != nil {
		return err
	}

	// Additional validation for runtime changes
	if newCfg.Concurrency < s.config.Concurrency {
		// Check if backup is in progress
		if s.metrics.IsBackupInProgress() {
			return newBackupError(
				"ValidateConfigChange",
				"",
				fmt.Errorf("cannot reduce concurrency while tasks are in progress"),
			)
		}
	}

	return nil
}
