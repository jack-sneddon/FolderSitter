package util

import (
	"fmt"

	"github.com/jack-sneddon/FolderSitter/internal/config"
)

// PrintUsage prints the backup plan, showing source-to-target mappings.
func PrintUsage(cfg *config.BackupConfig) {
	fmt.Println("\nBackup Plan:")
	for _, folder := range cfg.FoldersToBackup {
		sourcePath := fmt.Sprintf("%s/%s", cfg.SourceDirectory, folder)
		targetPath := fmt.Sprintf("%s/%s", cfg.TargetDirectory, folder)
		fmt.Printf("Copy from: %s -> To: %s\n", sourcePath, targetPath)
	}
	fmt.Printf("\n----------------\n")

	fmt.Println("All folders validated successfully. Ready to start the backup process.")
}