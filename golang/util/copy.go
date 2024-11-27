package util

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DeepCopy copies files and directories from source to target while preserving metadata and permissions.
func DeepCopy(source, target string, deepDuplicateCheck bool) error {
	// Start the timer for the entire DeepCopy process
	start := time.Now()

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
			fmt.Printf("Skipping %s, already up-to-date.\n", targetPath)
			return nil
		}

		return copyFile(path, targetPath, info)
	})

	// Print time taken for DeepCopy
	fmt.Printf("DeepCopy completed for %s in %s.\n", source, formatDuration(time.Since(start)))
	return err
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
