package backup

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Contains performCopy(), copyFile(), and related methods
// performCopy executes a single copy operation
func (s *Service) performCopy(task CopyTask) error {
	startTime := time.Now()

	src, err := os.Open(task.Source)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(task.Destination), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	dst, err := os.Create(task.Destination)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy with progress tracking and checksum calculation
	buf := make([]byte, s.config.BufferSize)
	hasher := sha256.New()
	writer := io.MultiWriter(dst, hasher)

	copied, err := io.CopyBuffer(writer, src, buf)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Calculate operation duration
	duration := time.Since(startTime)
	speedMBps := float64(copied) / 1024 / 1024 / duration.Seconds()

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

	s.logger.Info("Copied %s (%.2f MB) at %.2f MB/s",
		task.Source,
		float64(copied)/1024/1024,
		speedMBps)

	return nil
}

// Add this method to Service in backup.go
func (s *Service) copyFile(task CopyTask) error {
	// First check if we should skip this file
	if skip, err := s.shouldSkipFile(task); err != nil {
		return err
	} else if skip {
		return nil
	}
	return s.performCopy(task)
}
