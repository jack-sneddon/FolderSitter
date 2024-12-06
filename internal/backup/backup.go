package backup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jack-sneddon/FolderSitter/internal/config"
	"github.com/jack-sneddon/FolderSitter/internal/util"
)

var verbose bool
var quiet bool
var logFile *os.File

// SetLogLevel sets the logging level (info, warn, error).
func SetLogLevel(level string) {
	switch level {
	case "info":
		// Default logging behavior
		log.SetFlags(log.Ldate | log.Ltime)
	case "warn":
		// Suppress informational messages, only log warnings and errors
		log.SetFlags(log.Ldate | log.Ltime)
	case "error":
		// Suppress all messages except errors
		log.SetFlags(log.Ldate | log.Ltime)
	default:
		log.Printf("Invalid log level: %s. Defaulting to 'info'.", level)
		log.SetFlags(log.Ldate | log.Ltime)
	}
}

// initLogger initializes the logging system, redirecting logs to a file.
func initLogger() error {
	// Create "out" directory if it doesn't exist
	if err := os.MkdirAll("out", 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Generate log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFilePath := filepath.Join("out", fmt.Sprintf("out_%s.log", timestamp))

	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	// Redirect all logs to the file
	log.SetOutput(logFile)
	return nil
}

// closeLogger closes the log file.
func closeLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// SetVerbose toggles verbose logging.
func SetVerbose(v bool) {
	verbose = v
}

// SetQuiet toggles quiet mode.
func SetQuiet(q bool) {
	quiet = q
}

// logInfo logs informational messages to the log file only.
func logInfo(format string, args ...interface{}) {
	log.Printf("[INFO] "+format, args...)
}

// logError logs error messages to the log file and returns the message to the caller.
func logError(format string, args ...interface{}) {
	log.Printf("[ERROR] "+format, args...)
}

// logVerbose logs detailed debug messages to the log file in verbose mode.
func logVerbose(format string, args ...interface{}) {
	if verbose {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Run performs the backup process based on the provided configuration and journal file.
func Run(cfg *config.BackupConfig, journalFilePath string) error {
	if err := initLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer closeLogger()

	// Start spinner
	done := make(chan bool)
	go spinner(done)

	// Start timer
	startTime := time.Now()
	totalFoldersCopied := 0
	totalFoldersSkipped := 0

	// Ensure the target directory exists
	if err := os.MkdirAll(cfg.TargetDirectory, 0755); err != nil {
		logError("Failed to create target directory: %v", err)
		done <- true // Stop spinner
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	for _, folder := range cfg.FoldersToBackup {
		sourcePath := filepath.Join(cfg.SourceDirectory, folder)
		targetPath := filepath.Join(cfg.TargetDirectory, folder)

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			logError("Folder does not exist: %s", sourcePath)
			totalFoldersSkipped++
			continue
		}

		logInfo("Backing up folder: %s -> %s", sourcePath, targetPath)

		// Perform folder copy
		err := util.DeepCopy(sourcePath, targetPath, cfg.DeepDuplicateCheck, journalFilePath)
		if err != nil {
			logError("Failed to copy folder %s: %v", sourcePath, err)
			totalFoldersSkipped++
			continue
		}

		totalFoldersCopied++
	}

	// Stop spinner
	done <- true

	// Write journal entry
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	journalEntry := fmt.Sprintf(
		"===== Backup Run =====\nStart Time: %s\nEnd Time: %s\nDuration: %s\n"+
			"Folders Copied: %d, Folders Skipped: %d\n\n",
		startTime.Format("2006-01-02 15:04:05"),
		endTime.Format("2006-01-02 15:04:05"),
		duration.String(),
		totalFoldersCopied, totalFoldersSkipped,
	)
	if err := util.AppendToJournal(journalFilePath, journalEntry); err != nil {
		logError("Failed to write to journal: %v", err)
		return fmt.Errorf("failed to write to journal: %w", err)
	}

	// Print success or failure message to the terminal
	if totalFoldersCopied > 0 {
		fmt.Println("\nBackup completed successfully.")
	} else {
		fmt.Println("\nBackup completed with errors.")
	}
	return nil
}

// spinner displays a rotating spinner in the terminal to indicate progress.
func spinner(done chan bool) {
	spinChars := []rune{'|', '/', '-', '\\'}
	i := 0
	for {
		select {
		case <-done:
			fmt.Print("\r") // Clear spinner line
			return
		default:
			fmt.Printf("\rWorking... %c", spinChars[i])
			i = (i + 1) % len(spinChars)
			time.Sleep(200 * time.Millisecond) // Update spinner every 200ms
		}
	}
}

// DryRun simulates the backup process without making any changes.
func DryRun(cfg *config.BackupConfig, journalFilePath string) error {
	if err := initLogger(); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer closeLogger()

	logInfo("Starting Dry Run...")
	for _, folder := range cfg.FoldersToBackup {
		sourcePath := filepath.Join(cfg.SourceDirectory, folder)
		targetPath := filepath.Join(cfg.TargetDirectory, folder)

		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			logError("Folder does not exist: %s", sourcePath)
			continue
		}

		logInfo("[DRY-RUN] Would copy folder: %s -> %s", sourcePath, targetPath)
	}

	// Write a journal entry for the dry run
	journalEntry := fmt.Sprintf(
		"===== Dry Run =====\nStart Time: %s\nEnd Time: %s\n"+
			"Simulated folders to copy: %d\n\n",
		time.Now().Format("2006-01-02 15:04:05"),
		time.Now().Format("2006-01-02 15:04:05"),
		len(cfg.FoldersToBackup),
	)
	if err := util.AppendToJournal(journalFilePath, journalEntry); err != nil {
		logError("Failed to write to journal: %v", err)
		return fmt.Errorf("failed to write to journal: %w", err)
	}

	logInfo("Dry Run completed successfully.")
	return nil
}
