package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// DeepCopy copies files and directories from source to target while preserving metadata and permissions.
func DeepCopy(source, target string, deepDuplicateCheck bool, journalFilePath string) error {
	// Start the timer for the folder copy
	start := time.Now()

	// Initialize counters
	var copiedCount, skippedCount int

	// Traverse the source directory recursively
	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing %s: %v", path, err)
		}

		// Construct the corresponding target path
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("error calculating relative path: %v", err)
		}
		targetPath := filepath.Join(target, relPath)

		// Handle directories
		if info.IsDir() {
			if err := os.MkdirAll(targetPath, info.Mode()); err != nil {
				return fmt.Errorf("error creating directory %s: %v", targetPath, err)
			}
			return nil
		}

		// Handle files
		shouldCopy, err := shouldCopyFile(path, targetPath, info, deepDuplicateCheck)
		if err != nil {
			return fmt.Errorf("error comparing files: %v", err)
		}

		if !shouldCopy {
			skippedCount++
			return nil
		}

		if err := copyFile(path, targetPath, info); err != nil {
			return fmt.Errorf("error copying file: %v", err)
		}
		copiedCount++
		return nil
	})

	if err != nil {
		return err
	}

	// Log the folder summary
	folderSummary := fmt.Sprintf("Directory: %s\nCopied: %d files, Skipped: %d files\nElapsed Time: %s\n\n",
		source, copiedCount, skippedCount, formatDuration(time.Since(start)))
	if err := AppendToJournal(journalFilePath, folderSummary); err != nil {
		return fmt.Errorf("failed to write folder summary to journal: %v", err)
	}

	return nil
}

// shouldCopyFile determines whether a file needs to be copied based on metadata and optional checksum.
func shouldCopyFile(sourcePath, targetPath string, sourceInfo os.FileInfo, deepDuplicateCheck bool) (bool, error) {
	// Check if the target file exists
	targetInfo, err := os.Stat(targetPath)
	if os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("error stating target file %s: %v", targetPath, err)
	}

	// Compare size and permissions
	if sourceInfo.Size() != targetInfo.Size() || sourceInfo.Mode() != targetInfo.Mode() {
		return true, nil
	}

	// Compare checksum if enabled
	if deepDuplicateCheck {
		sourceChecksum, err := calculateChecksum(sourcePath)
		if err != nil {
			return false, fmt.Errorf("error calculating checksum for source file %s: %v", sourcePath, err)
		}
		targetChecksum, err := calculateChecksum(targetPath)
		if err != nil {
			return false, fmt.Errorf("error calculating checksum for target file %s: %v", targetPath, err)
		}
		if sourceChecksum != targetChecksum {
			return true, nil
		}
	}

	return false, nil
}

// copyFile performs the actual file copying while preserving metadata.
func copyFile(sourcePath, targetPath string, sourceInfo os.FileInfo) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("error opening source file %s: %v", sourcePath, err)
	}
	defer sourceFile.Close()

	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("error creating target file %s: %v", targetPath, err)
	}
	defer targetFile.Close()

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		return fmt.Errorf("error copying file content from %s to %s: %v", sourcePath, targetPath, err)
	}
	return nil
}

// calculateChecksum computes the SHA-256 checksum of a file.
func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening file %s: %v", filePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// AppendToJournal appends a message to the journal file.
func AppendToJournal(filePath, message string) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(message)
	return err
}

// formatDuration formats a time.Duration into a user-friendly string.
func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%02dh:%02dm:%02ds", d/time.Hour, (d%time.Hour)/time.Minute, (d%time.Minute)/time.Second)
}
