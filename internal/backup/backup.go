package backup

import (
	"fmt"
	"path/filepath"

	"github.com/jack-sneddon/FolderSitter/internal/config"
	"github.com/jack-sneddon/FolderSitter/internal/util"
)

// Run executes the entire backup process.
func Run(cfg *config.BackupConfig, journalFilePath string) error {
	// Step 1: Validate configuration
	if err := util.Validate(cfg); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Step 2: Prepare the journal file
	startMessage := fmt.Sprintf("\n\n===== Backup Run Started: %s =====\n", util.CurrentTimestamp())
	if err := util.AppendToJournal(journalFilePath, startMessage); err != nil {
		return fmt.Errorf("failed to write to journal: %w", err)
	}

	// Step 3: Execute the worker pool for folder copies
	util.WorkerPool(2, cfg.FoldersToBackup, func(folder string) {
		sourcePath := filepath.Join(cfg.SourceDirectory, folder)
		targetPath := filepath.Join(cfg.TargetDirectory, folder)
		fmt.Printf("\nBacking up %s to %s...\n", sourcePath, targetPath)

		// Perform the deep copy for the folder
		if err := util.DeepCopy(sourcePath, targetPath, cfg.DeepDuplicateCheck, journalFilePath); err != nil {
			fmt.Printf("Error processing folder %s: %v\n", folder, err)
		}
	})

	// Step 4: Log completion message to the journal
	endMessage := fmt.Sprintf("===== Backup Run Completed: %s =====\n", util.CurrentTimestamp())
	if err := util.AppendToJournal(journalFilePath, endMessage); err != nil {
		return fmt.Errorf("failed to write to journal: %w", err)
	}

	return nil
}