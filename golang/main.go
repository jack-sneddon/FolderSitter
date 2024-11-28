/*
My main usecase is backup from an SSD to HDD.
Because I'm using HDD, concurrency is kept to a minimum.

	- HDDs perform best with sequential writes, and excessive concurrency can degrade performance due to frequent seek operations.
	- Managing concurrency at the folder level (with sequential copying of files within folders) minimizes random writes to the HDD while leveraging the SSD's read speed.
	- The copying logic must handle recursive traversal of subdirectories, ensuring all files are backed up.

1. Worker Pool for Folder-Level Concurrency:
	- Use a worker pool to process folders concurrently, with a limited number of workers (e.g., 2–4 workers) to avoid overwhelming the HDD.
	- Each worker handles one folder at a time.

2. Sequential File Copying Within Folders:
	- Files in each folder (including subdirectories) are copied sequentially to maintain efficient HDD write patterns.

3. Recursive Copying:
	- Ensure the DeepCopy function recursively traverses and processes subdirectories.
*/

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

	// Read the configuration from the JSON file
	config, err := util.ReadConfig("backup_config.json")
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	// Validate the configuration directories and folders
	fmt.Println("Validating directories and folders...")
	if err := util.Validate(config); err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	// Print the backup plan
	util.PrintUsage(config)

	// Prepare the journal file path
	journalFilePath := filepath.Join(config.TargetDirectory, "folder-sitter-journal.txt")

	// Write the start of the run to the journal
	startMessage := fmt.Sprintf("\n\n===== Backup Run Started: %s =====\n", time.Now().Format("2006-01-02 15:04:05"))
	if err := util.AppendToJournal(journalFilePath, startMessage); err != nil {
		log.Fatalf("Failed to write to journal: %v", err)
	}

	// Start concurrent folder copies with a spinner
	fmt.Println("\nStarting the backup process...")
	done := make(chan bool) // Channel to signal the spinner
	go spinner(done)        // Start the spinner in a separate goroutine

	// Execute the worker pool for folder copies
	util.WorkerPool(2, config.FoldersToBackup, func(folder string) {
		sourcePath := filepath.Join(config.SourceDirectory, folder)
		targetPath := filepath.Join(config.TargetDirectory, folder)
		fmt.Printf("\nBacking up %s to %s...\n", sourcePath, targetPath)

		// Perform the deep copy for the folder
		if err := util.DeepCopy(sourcePath, targetPath, config.DeepDuplicateCheck, journalFilePath); err != nil {
			log.Printf("Error processing folder %s: %v", folder, err)
		}
	})

	// Stop the spinner
	done <- true

	// Calculate and log total time
	totalTime := formatDuration(time.Since(programStart))
	endMessage := fmt.Sprintf("===== Backup Run Completed: %s =====\nTotal Time: %s\n", time.Now().Format("2006-01-02 15:04:05"), totalTime)
	if err := util.AppendToJournal(journalFilePath, endMessage); err != nil {
		log.Fatalf("Failed to write to journal: %v", err)
	}

	// Print the total program time
	fmt.Printf("\nBackup process completed in %s.\n", totalTime)
}

// spinner displays a rotating spinner in the console to indicate progress.
func spinner(done chan bool) {
	spinChars := []rune{'|', '/', '-', '\\'}
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r") // Clear the spinner line
			return
		default:
			fmt.Printf("\rWorking... %c", spinChars[i])
			i = (i + 1) % len(spinChars)
			time.Sleep(200 * time.Millisecond) // Update spinner every 200ms
		}
	}
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
