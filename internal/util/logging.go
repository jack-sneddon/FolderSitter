package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	logger  *log.Logger
	logFile *os.File
)

var currentLogLevel string

// SetLogLevel sets the logging level for the application.
func SetLogLevel(level string) error {
	switch level {
	case "info":
		currentLogLevel = "info"
	case "warn":
		currentLogLevel = "warn"
	case "error":
		currentLogLevel = "error"
	default:
		return fmt.Errorf("invalid log level: %s", level)
	}
	return nil
}

// InitLogger initializes the logger to write to a file in the 'out' directory.
func InitLogger() error {
	// Ensure the 'out' directory exists
	if err := os.MkdirAll("out", 0755); err != nil {
		return err
	}

	// Create a log file with a timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFilePath := filepath.Join("out", "log-"+timestamp+".log")

	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Initialize the logger
	logger = log.New(logFile, "", log.LstdFlags)
	return nil
}

// CloseLogger closes the log file.
func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

// LogInfo logs informational messages if the log level is "info".
func LogInfo(format string, v ...interface{}) {
	if currentLogLevel == "info" {
		if logger != nil {
			logger.Printf("[INFO] "+format, v...)
		}
	}
}

// LogWarn logs warning messages if the log level is "info" or "warn".
func LogWarn(format string, v ...interface{}) {
	if currentLogLevel == "info" || currentLogLevel == "warn" {
		if logger != nil {
			logger.Printf("[WARN] "+format, v...)
		}
	}
}

// LogError logs error messages at all log levels.
func LogError(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, v...)
	}
}

// LogVerbose logs verbose messages to the log file.
func LogVerbose(verbose bool, format string, v ...interface{}) {
	if verbose && logger != nil {
		logger.Printf("[VERBOSE] "+format, v...)
	}
}

// AppendToJournal appends summary messages to a journal file.
func AppendToJournal(journalFilePath, message string) error {
	// Ensure the directory for the journal file exists
	dir := filepath.Dir(journalFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Open the journal file in append mode
	journalFile, err := os.OpenFile(journalFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer journalFile.Close()

	// Write the message to the journal file
	_, err = journalFile.WriteString(message + "\n")
	return err
}