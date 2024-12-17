package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jack-sneddon/FolderSitter/internal/backup"
)

func printHelp() {
	fmt.Print(`FolderSitter - Backup Utility

Usage:
  foldersitter [options]

Options:
  -config <file>       Path to the configuration file (JSON or YAML).
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

	// Create backup configuration
	cfg, err := backup.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Set configuration options from flags
	cfg.Options = &backup.Options{
		Verbose:  *verboseFlag,
		Quiet:    *quietFlag,
		LogLevel: *logLevel,
	}

	// Create backup service
	service, err := backup.NewService(cfg)
	if err != nil {
		fmt.Printf("Failed to create backup service: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration if requested
	if *validateFlag {
		if err := backup.Validate(cfg); err != nil {
			fmt.Printf("Configuration validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Configuration is valid.")
		return
	}

	// Create context for the operation
	ctx := context.Background()

	// Perform the operation
	if *dryRunFlag {
		if !*quietFlag {
			fmt.Println("Starting dry run...")
		}
		if err := service.DryRun(ctx); err != nil {
			fmt.Printf("Dry run failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		if !*quietFlag {
			fmt.Println("Starting backup...")
		}
		if err := service.Backup(ctx); err != nil {
			fmt.Printf("Backup failed: %v\n", err)
			os.Exit(1)
		}
	}

	if !*quietFlag {
		fmt.Println("Operation completed successfully.")
	}
}
