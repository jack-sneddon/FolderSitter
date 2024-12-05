package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/jack-sneddon/FolderSitter/internal/backup"
	"github.com/jack-sneddon/FolderSitter/internal/config"
)

func main() {
	// Parse command-line arguments
	configPath := flag.String("config", "", "Path to the configuration file")
	flag.Parse()

	if *configPath == "" {
		log.Fatalf("Usage: foldersitter -config <path_to_config>")
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Determine the journal file path
	journalFilePath := filepath.Join(cfg.TargetDirectory, "folder-sitter-journal.txt")

	// Run the backup process
	if err := backup.Run(cfg, journalFilePath); err != nil {
		log.Fatalf("Backup process failed: %v", err)
	}

	log.Println("Backup completed successfully.")
}