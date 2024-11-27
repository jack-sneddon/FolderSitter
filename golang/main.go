package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/jack-sneddon/FolderSitter/golang/util"
)

func main() {
	// Start the program timer
	programStart := time.Now()

	// 1. Read the configuration from JSON
	config, err := util.ReadConfig("backup_config.json")
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	// 2. Validate the configuration directories and folders
	fmt.Println("Validating directories and folders...")
	if err := util.Validate(config); err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	// 3. Print the backup plan
	util.PrintUsage(config)

	// 4. Perform the copy for each folder
	fmt.Println("Starting the backup process...")
	for _, folder := range config.FoldersToBackup {
		sourcePath := filepath.Join(config.SourceDirectory, folder)
		targetPath := filepath.Join(config.TargetDirectory, folder)
		fmt.Printf("Backing up %s to %s...\n", sourcePath, targetPath)

		// Start the timer for DeepCopy
		copyStart := time.Now()
		if err := util.DeepCopy(sourcePath, targetPath, config.DeepDuplicateCheck); err != nil {
			log.Printf("Error copying %s: %v", folder, err)
		}
		// Print time taken for this folder
		fmt.Printf("Completed copy of %s in %s.\n", folder, formatDuration(time.Since(copyStart)))
	}

	// 5. Validate the copied data
	fmt.Println("Validating the copied data...")
	for _, folder := range config.FoldersToBackup {
		sourcePath := filepath.Join(config.SourceDirectory, folder)
		targetPath := filepath.Join(config.TargetDirectory, folder)
		fmt.Printf("Validating copy of %s...\n", folder)
		if err := util.ValidateCopy(sourcePath, targetPath); err != nil {
			log.Printf("Validation failed for %s: %v", folder, err)
		} else {
			fmt.Printf("Validation successful for %s.\n", folder)
		}
	}

	// Print the total program time
	fmt.Printf("\nProgram completed in %s.\n", formatDuration(time.Since(programStart)))
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
