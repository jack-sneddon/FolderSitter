package util

import (
	"fmt"
	"os"

	"github.com/jack-sneddon/FolderSitter/internal/config"
)

// Validate ensures the backup configuration is valid.
func Validate(cfg *config.BackupConfig) error {
	// Check source directory
	if _, err := os.Stat(cfg.SourceDirectory); os.IsNotExist(err) {
		return fmt.Errorf("source directory %s does not exist", cfg.SourceDirectory)
	}

	// Ensure target directory exists or create it
	if _, err := os.Stat(cfg.TargetDirectory); os.IsNotExist(err) {
		if err := os.MkdirAll(cfg.TargetDirectory, 0755); err != nil {
			return fmt.Errorf("unable to create target directory %s: %w", cfg.TargetDirectory, err)
		}
		fmt.Printf("Target directory %s created.\n", cfg.TargetDirectory)
	}

	// Validate folders to back up
	for _, folder := range cfg.FoldersToBackup {
		folderPath := fmt.Sprintf("%s/%s", cfg.SourceDirectory, folder)
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			fmt.Printf("Warning: Folder %s does not exist, skipping...\n", folder)
		}
	}

	return nil
}
