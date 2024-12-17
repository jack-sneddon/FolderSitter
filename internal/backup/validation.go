package backup

import (
	"fmt"
	"os"
)

// Contains validatePaths(), shouldSkipFile(), and related methods

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
	s.metrics.mu.Lock()
	s.metrics.BytesCopied += sourceInfo.Size() // Include size in total
	s.metrics.FilesSkipped++
	s.metrics.mu.Unlock()

	s.logger.Debug("Skipped identical file: %s (Size: %.2f MB)",
		task.Source, float64(sourceInfo.Size())/1024/1024)
	return true, nil
}