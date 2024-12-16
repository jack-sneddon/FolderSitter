package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jack-sneddon/FolderSitter/internal/config"
	"github.com/jack-sneddon/FolderSitter/internal/util"
)

func printHelp() {
	fmt.Println(`FolderSitter - Backup Utility

Usage:
  foldersitter [options]

Options:
  -config <file>       Path to the configuration file (JSON or YAML).
  --help, -h           Show this help message and exit.
  --verbose, -v        Enable verbose logging.
  --quiet, -q          Suppress all output except errors.
  --validate           Validate the configuration file without performing a backup.
  --dry-run            Simulate the backup process without making any changes.
  --log-level <level>  Set logging level: info, warn, error.

Examples:
  foldersitter -config backup_config.json
  foldersitter -config backup_config.yaml --dry-run --verbose
`)
}

func main() {
	// Parse CLI flags
	configPath := flag.String("config", "", "Path to the configuration file")
	helpFlag := flag.Bool("help", false, "Show help message")
	verboseFlag := flag.Bool("verbose", false, "Enable verbose logging")
	quietFlag := flag.Bool("quiet", false, "Suppress all output except errors")
	validateFlag := flag.Bool("validate", false, "Validate the configuration file without performing a backup")
	dryRunFlag := flag.Bool("dry-run", false, "Simulate the backup process without making any changes")
	logLevel := flag.String("log-level", "info", "Set logging level: info, warn, error")
	flag.Parse()

	// Show help message if --help or -h is provided
	if *helpFlag {
		printHelp()
		return
	}

	// Validate required flags
	if *configPath == "" {
		fmt.Println("Error: -config flag is required.")
		printHelp()
		os.Exit(1)
	}

	// Initialize logger
	if err := util.InitLogger(); err != nil {
		fmt.Println("Failed to initialize logger:", err)
		os.Exit(1)
	}
	defer util.CloseLogger()

	// Set log level
	if err := util.SetLogLevel(*logLevel); err != nil {
		fmt.Println("Invalid log level:", *logLevel)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		util.LogError("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	// Validate configuration if --validate is provided
	if *validateFlag {
		if err := config.Validate(cfg); err != nil {
			util.LogError("Configuration validation failed: %v", err)
			os.Exit(1)
		}
		fmt.Println("Configuration is valid.")
		return
	}

	// Perform backup
	if *dryRunFlag {
		util.LogInfo("Performing dry run...")
	} else {
		util.LogInfo("Starting backup...")
	}

	// Define the journal file path
	journalFilePath := filepath.Join(cfg.TargetDirectory, "folder-sitter-journal.txt")

	for _, folder := range cfg.FoldersToBackup {
		srcPath := filepath.Join(cfg.SourceDirectory, folder)
		destPath := filepath.Join(cfg.TargetDirectory, folder)

		if *dryRunFlag {
			util.LogInfo("Dry run: would copy %s to %s", srcPath, destPath)
		} else {
			if err := util.CopyDirectory(srcPath, destPath, cfg.DeepDuplicateCheck, journalFilePath); err != nil {
				util.LogError("Failed to copy %s to %s: %v", srcPath, destPath, err)
				if !*quietFlag {
					fmt.Printf("Error: failed to copy %s to %s\n", srcPath, destPath)
				}
			} else {
				util.LogInfo("Successfully copied %s to %s", srcPath, destPath)
				if !*quietFlag && !*verboseFlag {
					fmt.Printf("Copied %s to %s\n", srcPath, destPath)
				}
			}
		}
	}

	util.LogInfo("Backup process completed.")
	if !*quietFlag {
		fmt.Println("Backup process completed.")
	}
}