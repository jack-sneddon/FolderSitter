package util

import (
	"fmt"
	"os"
	"time"
)

// AppendToJournal writes a log entry to the specified journal file.
func AppendToJournal(journalFilePath, message string) error {
	// Open the journal file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(journalFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open journal file: %w", err)
	}
	defer file.Close()

	// Write the message to the journal
	if _, err := file.WriteString(message + "\n"); err != nil {
		return fmt.Errorf("failed to write to journal: %w", err)
	}

	return nil
}

// LogInfo formats and writes an informational log entry.
func LogInfo(journalFilePath, message string) error {
	timestamp := CurrentTimestamp()
	formattedMessage := fmt.Sprintf("[INFO] %s: %s", timestamp, message)
	fmt.Println(formattedMessage) // Print to console
	return AppendToJournal(journalFilePath, formattedMessage)
}

// LogError formats and writes an error log entry.
func LogError(journalFilePath, message string) error {
	timestamp := CurrentTimestamp()
	formattedMessage := fmt.Sprintf("[ERROR] %s: %s", timestamp, message)
	fmt.Fprintln(os.Stderr, formattedMessage) // Print to stderr
	return AppendToJournal(journalFilePath, formattedMessage)
}

// CurrentTimestamp returns the current time as a formatted string.
func CurrentTimestamp() string {
    return time.Now().Format("2006-01-02 15:04:05")
}