# backup-butler

A versatile and customizable backup utility designed to ensure safe, efficient, and reliable backups of your critical data. **backup-butler** provides features like checksum validation, logging, and flexible configuration to meet various backup needs.

## Features
- **Customizable Backup Configurations**: Define source directories and folders to back up via a configuration file.
- **Integrity Checks**: Perform checksums for individual files and entire directories to ensure data consistency.
- **Detailed Logs**: Maintain a backup log with results of the copy operation and checksum validations.
- **Incremental Backups**: Copy only files that have changed.
- **Optional Compression**: Save space by compressing backed-up files.
- **Exclusion Rules**: Exclude specific file types or patterns from backups.
- **Scheduling**: Automate backups using your system's task scheduler (e.g., cron for Linux, Task Scheduler for Windows).
- **Restore Functionality**: Optionally include functionality to restore files from a backup.

## Requirements
- A configuration file with the following structure:
    ```json
    {
      "source_directory": "/path/to/source",
      "folders_to_backup": [
        "folder1",
        "folder2",
        "folder3/subfolder"
      ]
    }
    ```
- A runtime environment for the selected implementation (e.g., Python, Go, or a shell interpreter).

## Usage
1. Clone the repository:
    ```bash
    git clone https://github.com/jack-sneddon/backup-butler.git
    cd backup-butler
    ```

2. Prepare your configuration file:
    - Create a configuration file (e.g., `backup_config.json`) as described in the **Requirements** section.

3. Execute the utility:
    - Depending on the implementation, run the utility with the appropriate command:
        - **Python**:
            ```bash
            python foldersitter.py --config backup_config.json --destination /path/to/destination
            ```
        - **Go**:
            ```bash
            go run backup-butler.go --config backup_config.json --destination /path/to/destination
            ```
        - **Shell Script**:
            ```bash
            ./backup-butler.sh --config backup_config.json --destination /path/to/destination
            ```

4. Review the backup log:
    - A detailed log file will be generated in the specified destination folder.

## Changelog
### Version 1.0.0
- Initial release.
- Support for customizable backup configurations.
- File and directory checksum validation.
- Basic logging functionality.

## Future Enhancements
- Support for remote destinations (e.g., AWS S3, Google Drive).
- Encryption for sensitive data during backups.
- Multi-threaded or parallel processing for improved performance.
- Real-time progress reporting.
- Cross-platform GUI for easier management.
- Backup retention policies to manage storage.
- Pre/post-backup hooks for custom actions.
- Notifications for success, failure, or errors during backup.
- Cloud-based configuration storage for syncing across systems.

## Suggestions for Future Features
	•	Support for Remote Destinations: Implement functionality to back up data to remote destinations such as AWS S3 or Google Drive, providing users with more flexibility in storage options.
	•	Encryption for Sensitive Data: Add encryption features to secure sensitive data during backups, ensuring data privacy and protection.
	•	Multi-threaded or Parallel Processing: Introduce multi-threaded or parallel processing capabilities to improve performance, especially when dealing with large datasets.
	•	Real-time Progress Reporting: Provide real-time progress reporting to keep users informed about the backup process status.
	•	Cross-platform GUI: Develop a cross-platform graphical user interface (GUI) to make the tool more accessible to users who prefer not to use command-line interfaces.
	•	Backup Retention Policies: Implement backup retention policies to manage storage by automatically deleting old backups based on user-defined criteria.
	•	Pre/Post-backup Hooks: Allow users to define custom actions or scripts to be executed before and after the backup process, providing greater flexibility in backup operations.
	•	Notifications: Set up notifications to inform users of the success, failure, or errors during the backup process, enhancing user awareness and prompt action.
	•	Cloud-based Configuration Storage: Enable cloud-based storage for configuration files to allow syncing across multiple systems, ensuring consistency in backup settings.

# code structure
.
├── cmd/
│   ├── cli/
│   │   └── main.go               # CLI entry point
│   └── gui/
│       └── main.go               # GUI entry point
├── internal/
│   ├── domain/                   # Core business logic & interfaces
│   │   ├── backup/
│   │   │   ├── service.go        # Core backup service interface
│   │   │   ├── types.go          # Domain models and types
│   │   │   └── ports.go          # Interface definitions for backup
│   │   └── versioning/
│   │       ├── manager.go        # Version management interface
│   │       ├── types.go          # Version-related types
│   │       └── ports.go          # Interface definitions for versioning
│   ├── core/                     # Use cases & shared services
│   │   ├── backup/
│   │   │   ├── service.go        # Implementation of backup service
│   │   │   ├── operations.go     # Backup operations implementation
│   │   │   └── validator.go      # Backup validation logic
│   │   ├── storage/
│   │   │   ├── manager.go        # Storage coordination
│   │   │   ├── copy.go          # File copy operations (from your existing)
│   │   │   ├── checksum.go      # Checksum operations (from your existing)
│   │   │   └── errors.go        # Storage-specific errors
│   │   ├── monitoring/
│   │   │   ├── logger.go        # Logger implementation (from your existing)
│   │   │   ├── metrics.go       # Metrics implementation (from your existing)
│   │   │   └── types.go         # Monitoring types
│   │   ├── config/
│   │   │   ├── loader.go        # Config loading logic (from your existing)
│   │   │   ├── validator.go     # Config validation (from your existing)
│   │   │   └── types.go         # Configuration types
│   │   └── worker/
│   │       ├── pool.go          # Worker pool implementation (from your existing)
│   │       └── task.go          # Task definitions (from your existing)
│   ├── adapters/                # Implementation of ports
│   │   ├── storage/
│   │   │   ├── filesystem/
│   │   │   │   ├── adapter.go   # Filesystem storage implementation
│   │   │   │   └── helper.go    # Filesystem utility functions
│   │   │   └── mock/
│   │   │       └── adapter.go   # Mock storage for testing
│   │   └── metrics/
│   │       ├── prometheus/
│   │       │   └── adapter.go   # Prometheus metrics implementation
│   │       └── console/
│   │           └── adapter.go   # Console metrics implementation
│   └── ui/
│       ├── shared/              # Shared UI logic
│       │   ├── viewmodels/
│       │   │   ├── backup_vm.go  # Backup view model
│       │   │   └── config_vm.go  # Config view model
│       │   ├── controllers/
│       │   │   ├── backup.go     # Shared backup controller
│       │   │   └── config.go     # Shared config controller
│       │   └── events/
│       │       └── types.go      # Shared event definitions
│       ├── cli/
│       │   ├── commands/
│       │   │   ├── backup.go     # Backup command
│       │   │   ├── config.go     # Config command
│       │   │   └── version.go    # Version command
│       │   └── formatter/
│       │       └── output.go     # CLI output formatting
│       └── gui/
│           ├── views/
│           │   ├── backup/
│           │   │   ├── main.go    # Main backup view
│           │   │   ├── progress.go # Progress view
│           │   │   └── summary.go  # Summary view
│           │   └── config/
│           │       ├── editor.go   # Config editor view
│           │       └── validator.go # Config validation view
│           ├── widgets/
│           │   ├── file_picker.go  # Custom file picker
│           │   └── progress_bar.go # Custom progress bar
│           └── state/
│               └── store.go        # GUI state management
├── pkg/                         # Public API (if needed)
│   └── backup/
│       ├── client.go           # Client API
│       └── types.go            # Public types
├── configs/                    # Configuration files
│   ├── default.yaml
│   └── templates/
├── assets/                     # GUI assets
│   ├── icons/
│   └── styles/
└── tests/                      # Integration tests
    ├── backup_test.go
    └── fixtures/
    
## License
MIT License