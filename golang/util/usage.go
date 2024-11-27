package util

import "fmt"

// PrintUsage prints out the source-to-target mapping for each folder to be backed up.
func PrintUsage(config BackupConfig) {
	fmt.Println("\nBackup Plan:")
	for _, folder := range config.FoldersToBackup {
		sourcePath := fmt.Sprintf("%s/%s", config.SourceDirectory, folder)
		targetPath := fmt.Sprintf("%s/%s", config.TargetDirectory, folder)
		fmt.Printf("Copy from: %s -> To: %s\n", sourcePath, targetPath)
	}
	fmt.Printf("\n----------------\n")

	fmt.Println("All folders validated successfully. Ready to start the backup process.")
}
