package util

import (
	"fmt"
	"os"
)

// Validate checks the backup configuration, ensuring source and target directories exist and folders are valid.
func Validate(config BackupConfig) error {
	// Validate source directory
	if _, err := os.Stat(config.SourceDirectory); os.IsNotExist(err) {
		return fmt.Errorf("source directory %s does not exist", config.SourceDirectory)
	}

	// Validate target directory and create it if necessary
	if _, err := os.Stat(config.TargetDirectory); os.IsNotExist(err) {
		err := os.MkdirAll(config.TargetDirectory, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to create target directory %s: %v", config.TargetDirectory, err)
		}
		fmt.Printf("Target directory %s created.\n", config.TargetDirectory)
	}

	// Validate folders to backup
	for _, folder := range config.FoldersToBackup {
		folderPath := fmt.Sprintf("%s/%s", config.SourceDirectory, folder)
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			fmt.Printf("Warning: Folder %s does not exist, skipping...\n", folderPath)
		}
	}

	return nil
}
