package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jack-sneddon/FolderSitter/internal/backup"
)

func printHelp() {
	fmt.Print(`FolderSitter - Backup Utility

Usage:
  foldersitter [options]

Options:
  -config <file>       Path to the configuration file (JSON or YAML)
  --help, -h          Show this help message and exit
  --verbose, -v       Enable verbose logging
  --quiet, -q         Suppress all output except errors
  --validate          Validate the configuration file without performing a backup
  --dry-run           Simulate the backup process without making any changes
  --log-level <level> Set logging level: info, warn, error
  --list-versions     List all backup versions
  --show-version <id> Show details of a specific backup version
  --latest-version    Show most recent backup details

Examples:
  foldersitter -config backup_config.json
  foldersitter -config backup_config.yaml --dry-run --verbose
  foldersitter -config backup_config.yaml --list-versions
  foldersitter -config backup_config.yaml --show-version 20240117-150405
  foldersitter -config backup_config.yaml --latest-version
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
	listVersions := flag.Bool("list-versions", false, "List all backup versions")
	showVersion := flag.String("show-version", "", "Show details of a specific backup version")
	latestVersion := flag.Bool("latest-version", false, "Show most recent backup details")

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

	// Handle version management flags
	if *listVersions {
		printVersionList(service)
		return
	}
	if *showVersion != "" {
		printVersionDetails(service, *showVersion)
		return
	}
	if *latestVersion {
		version, err := service.GetLatestVersion()
		if err != nil {
			fmt.Printf("Error getting latest version: %v\n", err)
			return
		}
		printVersionDetails(service, version.ID)
		return
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

func printVersionList(service *backup.Service) {
	versions := service.GetVersions()
	if len(versions) == 0 {
		fmt.Println("No backup versions found")
		return
	}

	fmt.Println("\nBackup History:")
	fmt.Println("---------------")
	for _, v := range versions {
		fmt.Printf("ID: %s\n", v.ID)
		fmt.Printf("  Time: %s\n", v.Timestamp.Format(time.RFC3339))
		fmt.Printf("  Duration: %v\n", v.Duration)
		fmt.Printf("  Files: %d total (%d copied, %d skipped, %d failed)\n",
			v.Stats.TotalFiles, v.Stats.FilesBackedUp, v.Stats.FilesSkipped, v.Stats.FilesFailed)
		fmt.Printf("  Size: %.2f MB\n", float64(v.Size)/1024/1024)
		fmt.Printf("  Status: %s\n", v.Status)
		fmt.Println("---------------")
	}
}

func printVersionDetails(service *backup.Service, id string) {
	version, err := service.GetVersion(id)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nBackup Version Details: %s\n", version.ID)
	fmt.Printf("-------------------------\n")
	fmt.Printf("Timestamp: %s\n", version.Timestamp.Format(time.RFC3339))
	fmt.Printf("Duration: %v\n", version.Duration)
	fmt.Printf("Status: %s\n", version.Status)

	fmt.Printf("\nStatistics:\n")
	fmt.Printf("  Total Files Processed: %d\n", version.Stats.TotalFiles)
	fmt.Printf("  Files Backed Up: %d\n", version.Stats.FilesBackedUp)
	fmt.Printf("  Files Skipped: %d\n", version.Stats.FilesSkipped)
	fmt.Printf("  Files Failed: %d\n", version.Stats.FilesFailed)
	fmt.Printf("  Total Size: %.2f MB\n", float64(version.Stats.TotalBytes)/1024/1024)
	fmt.Printf("  Data Transferred: %.2f MB\n", float64(version.Stats.BytesTransferred)/1024/1024)

	fmt.Printf("\nConfiguration Used:\n")
	fmt.Printf("  Source Directory: %s\n", version.ConfigUsed.SourceDirectory)
	fmt.Printf("  Target Directory: %s\n", version.ConfigUsed.TargetDirectory)
	fmt.Printf("  Concurrency: %d\n", version.ConfigUsed.Concurrency)
	fmt.Printf("  Deep Duplicate Check: %v\n", version.ConfigUsed.DeepDuplicateCheck)
}
