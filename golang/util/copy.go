package util

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MaxBufferSize defines the maximum size of the in-memory buffer before flushing to the journal.
const MaxBufferSize = 10000 // 10,000 characters

// DeepCopy copies files and directories from source to target while preserving metadata and permissions.
func DeepCopy(source, target string, deepDuplicateCheck bool, journalFilePath string) error {
	// Start the timer for the entire DeepCopy process
	start := time.Now()

	// Initialize the buffer for the journal
	var journalBuffer strings.Builder

	// Add a journal entry header
	entryHeader := fmt.Sprintf("\n\n----- %s -----\n", time.Now().Format("2006-01-02 15:04:05"))
	journalBuffer.WriteString(entryHeader)

	// Counters for copied and skipped files
	var copiedCount, skippedCount int

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

		// Check if it's a directory
		if info.IsDir() {
			if err := os.MkdirAll(targetPath, info.Mode()); err != nil {
				return fmt.Errorf("error creating directory %s: %v", targetPath, err)
			}
			return nil
		}

		// If it's a file, check if it needs copying
		if shouldCopy, err := shouldCopyFile(path, targetPath, info, deepDuplicateCheck); err != nil {
			return fmt.Errorf("error comparing files: %v", err)
		} else if !shouldCopy {
			skippedCount++
		} else {
			if err := copyFile(path, targetPath, info); err != nil {
				return fmt.Errorf("error copying file: %v", err)
			}
			copiedCount++
		}

		// Flush the buffer periodically if it exceeds the max size
		if journalBuffer.Len() > MaxBufferSize {
			if err := flushJournalBuffer(journalFilePath, &journalBuffer); err != nil {
				return fmt.Errorf("error flushing journal buffer: %v", err)
			}
		}

		return nil
	})

	// Add a directory-specific summary
	summaryEntry := fmt.Sprintf("Directory: %s\nSummary: %d files copied, %d files skipped.\n\n", source, copiedCount, skippedCount)
	journalBuffer.WriteString(summaryEntry)

	// Add a footer with the total time taken
	durationEntry := fmt.Sprintf("Total time: %s.\n", formatDuration(time.Since(start)))
	journalBuffer.WriteString(durationEntry)

	// Final flush of the remaining buffer
	if flushErr := flushJournalBuffer(journalFilePath, &journalBuffer); flushErr != nil {
		return fmt.Errorf("error flushing final journal buffer: %v", flushErr)
	}

	return err
}

// flushJournalBuffer writes the current content of the journal buffer to the journal file and clears the buffer.
func flushJournalBuffer(filePath string, buffer *strings.Builder) error {
	if buffer.Len() == 0 {
		return nil // Nothing to flush
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the buffer content to the file
	if _, err := file.WriteString(buffer.String()); err != nil {
		return err
	}

	// Clear the buffer
	buffer.Reset()
	return nil
}

// shouldCopyFile checks if the file at the target path is identical to the source.
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
	// Open the source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("error opening source file %s: %v", sourcePath, err)
	}
	defer sourceFile.Close()

	// Create the target file with the same permissions as the source
	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("error creating target file %s: %v", targetPath, err)
	}
	defer targetFile.Close()

	// Copy the file content from source to target
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

// formatDuration formats a time.Duration into a user-friendly string.
func formatDuration(d time.Duration) string {
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second
	return fmt.Sprintf("%02dh:%02dm:%02ds", hours, minutes, seconds)
}
