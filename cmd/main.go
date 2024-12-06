package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jack-sneddon/FolderSitter/internal/backup"
	"github.com/jack-sneddon/FolderSitter/internal/config"
)

func printHelp() {
	fmt.Println(`FolderSitter - Backup Utility
Usage:
  foldersitter [options]

Options:
  -config <file>      Path to the configuration file (JSON or YAML).
  --help, -h          Show this help message and exit.
  --verbose, -v       Enable verbose logging.
  --quiet, -q         Suppress all output except errors.
  --validate          Validate the configuration file without performing a backup.
  --dry-run           Simulate the backup process without making any changes.
  --log-level <level> Set logging level: info, warn, error.

Examples:
  foldersitter -config backup_config.json
  foldersitter -config backup_config.yaml --dry-run --verbose
`)
}

func main() {
	// Parse CLI flags
	configPath := flag.String("config", "", "Path to the configuration file")
	helpFlag := flag.Bool("help", false, "Show help")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging")
	quietFlag := flag.Bool("quiet", false, "Suppress all output except errors")
	validateFlag := flag.Bool("validate", false, "Validate the configuration file")
	dryRunFlag := flag.Bool("dry-run", false, "Simulate the backup process without making changes")
	logLevel := flag.String("log-level", "info", "Set logging level: info, warn, error")

	flag.Parse()

	if *helpFlag || *configPath == "" {
		printHelp()
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Determine the journal file path
	journalFilePath := filepath.Join(cfg.TargetDirectory, "folder-sitter-journal.txt")

	// Handle --validate
	if *validateFlag {
		fmt.Println("Validating configuration...")
		if err := config.Validate(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration is valid.")
		os.Exit(0)
	}

	// Handle --dry-run
	if *dryRunFlag {
		fmt.Println("Dry run: Simulating backup process...")
		if err := backup.DryRun(cfg, journalFilePath); err != nil {
			log.Fatalf("Dry Run failed: %v", err)
		}
		os.Exit(0)
	}

	// Adjust logging based on flags
	backup.SetLogLevel(*logLevel)
	backup.SetVerbose(*verboseFlag)
	backup.SetQuiet(*quietFlag)

	// Perform backup
	if err := backup.Run(cfg, journalFilePath); err != nil {
		log.Fatalf("Backup failed: %v", err)
	}

	fmt.Println("Backup completed successfully.")
}
