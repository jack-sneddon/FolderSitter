# FolderSitter

A versatile and customizable backup utility designed to ensure safe, efficient, and reliable backups of your critical data. **FolderSitter** provides features like checksum validation, logging, and flexible configuration to meet various backup needs.

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
    git clone https://github.com/jack-sneddon/foldersitter.git
    cd foldersitter
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
            go run foldersitter.go --config backup_config.json --destination /path/to/destination
            ```
        - **Shell Script**:
            ```bash
            ./foldersitter.sh --config backup_config.json --destination /path/to/destination
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


## License
MIT License