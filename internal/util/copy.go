package util

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DeepCopy recursively copies a directory or file from source to destination.
func DeepCopy(sourcePath, destPath string, deepDuplicateCheck bool, journalFilePath string) error {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to access source path: %w", err)
	}

	// If source is a directory, copy recursively
	if sourceInfo.IsDir() {
		return copyDirectory(sourcePath, destPath, deepDuplicateCheck, journalFilePath)
	}

	// Otherwise, copy a single file
	return copyFile(sourcePath, destPath, deepDuplicateCheck, journalFilePath)
}

// copyDirectory copies all files and subdirectories from source to destination.
func copyDirectory(sourceDir, destDir string, deepDuplicateCheck bool, journalFilePath string) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("failed to read source directory %s: %w", sourceDir, err)
	}

	// Create the destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", destDir, err)
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDir, entry.Name())
		destPath := filepath.Join(destDir, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(sourcePath, destPath, deepDuplicateCheck, journalFilePath); err != nil {
				return err
			}
		} else {
			if err := copyFile(sourcePath, destPath, deepDuplicateCheck, journalFilePath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from source to destination.
func copyFile(sourceFile, destFile string, deepDuplicateCheck bool, journalFilePath string) error {
	// If deepDuplicateCheck is enabled, skip identical files
	if deepDuplicateCheck {
		isDuplicate, err := shouldSkipFile(sourceFile, destFile)
		if err != nil {
			return fmt.Errorf("failed to check file duplication: %w", err)
		}
		if isDuplicate {
			message := fmt.Sprintf("Skipped identical file: %s", sourceFile)
			LogInfo(journalFilePath, message)
			return nil
		}
	}

	// Open source file
	src, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy contents
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Preserve permissions
	sourceInfo, err := os.Stat(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}
	if err := os.Chmod(destFile, sourceInfo.Mode()); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Log the copy operation
	message := fmt.Sprintf("Copied file: %s -> %s", sourceFile, destFile)
	return LogInfo(journalFilePath, message)
}

// shouldSkipFile checks if the destination file is identical to the source file.
func shouldSkipFile(sourceFile, destFile string) (bool, error) {
	sourceInfo, err := os.Stat(sourceFile)
	if err != nil {
		return false, nil // Assume file doesn't exist
	}

	destInfo, err := os.Stat(destFile)
	if os.IsNotExist(err) {
		return false, nil // Destination file doesn't exist
	} else if err != nil {
		return false, err
	}

	// Compare file sizes and modification times
	return sourceInfo.Size() == destInfo.Size() && sourceInfo.ModTime() == destInfo.ModTime(), nil
}
