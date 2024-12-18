/*
When a backup runs:

Creates a .versions directory in your target backup location
For each backup run, creates a JSON file like 20240117-150405.json containing:

 - Timestamp of the backup
 - List of all files backed up
 - File metadata (sizes, checksums, modification times)
 - Statistics about the backup operation
 - Configuration used for the backup
 - Success/failure status

Benefits:

 - Track changes over time
 - Restore from specific points in time
 - Identify when files were modified
 - Detect files that were deleted
 - Calculate storage growth over time
*/

// version.go
package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupVersion represents a single backup operation
type BackupVersion struct {
	ID         string                  // Unique identifier (timestamp-based)
	Timestamp  time.Time               // When backup was performed
	Files      map[string]FileMetadata // Map of path to file metadata
	Size       int64                   // Total size of backup
	Status     string                  // Success, Failed, Partial
	Duration   time.Duration           // How long the backup took
	Stats      BackupStats             // Additional statistics
	ConfigUsed Config                  // Configuration used for this backup
}

// BackupStats holds statistical information about the backup
type BackupStats struct {
	TotalFiles       int   // Total number of files processed
	FilesBackedUp    int   // Number of files actually copied
	FilesSkipped     int   // Number of unchanged files
	FilesFailed      int   // Number of files that failed to backup
	TotalBytes       int64 // Total bytes processed
	BytesTransferred int64 // Actual bytes copied
}

// VersionManager handles backup versioning
type VersionManager struct {
	baseDir    string          // Base directory for version storage
	versions   []BackupVersion // List of all versions
	currentVer *BackupVersion  // Current backup version being processed
}

func NewVersionManager(baseDir string) (*VersionManager, error) {
	vm := &VersionManager{
		baseDir: baseDir,
	}

	// Create versions directory if it doesn't exist
	versionsDir := filepath.Join(baseDir, ".versions")
	if err := os.MkdirAll(versionsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create versions directory: %w", err)
	}

	// Load existing versions
	if err := vm.loadVersions(); err != nil {
		return nil, err
	}

	return vm, nil
}

func (vm *VersionManager) StartNewVersion(cfg *Config) *BackupVersion {
	version := &BackupVersion{
		ID:         time.Now().Format("20060102-150405"),
		Timestamp:  time.Now(),
		Files:      make(map[string]FileMetadata),
		Status:     "In Progress",
		ConfigUsed: *cfg,
	}
	vm.currentVer = version
	return version
}

func (vm *VersionManager) AddFile(path string, metadata FileMetadata) {
	if vm.currentVer != nil {
		vm.currentVer.Files[path] = metadata
		vm.currentVer.Size += metadata.Size
		vm.currentVer.Stats.TotalFiles++
		vm.currentVer.Stats.TotalBytes += metadata.Size
	}
}

func (vm *VersionManager) CompleteVersion(stats BackupStats) error {
	if vm.currentVer == nil {
		return fmt.Errorf("no backup version in progress")
	}

	vm.currentVer.Status = "Completed"
	vm.currentVer.Duration = time.Since(vm.currentVer.Timestamp)
	vm.currentVer.Stats = stats

	// Save the version
	if err := vm.saveVersion(vm.currentVer); err != nil {
		return err
	}

	vm.versions = append(vm.versions, *vm.currentVer)
	vm.currentVer = nil

	return nil
}

func (vm *VersionManager) saveVersion(ver *BackupVersion) error {
	filename := filepath.Join(vm.baseDir, ".versions", ver.ID+".json")

	data, err := json.MarshalIndent(ver, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal version data: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to save version file: %w", err)
	}

	return nil
}

func (vm *VersionManager) loadVersions() error {
	versionsDir := filepath.Join(vm.baseDir, ".versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read versions directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			filename := filepath.Join(versionsDir, entry.Name())
			data, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("failed to read version file %s: %w", entry.Name(), err)
			}

			var version BackupVersion
			if err := json.Unmarshal(data, &version); err != nil {
				return fmt.Errorf("failed to parse version file %s: %w", entry.Name(), err)
			}

			vm.versions = append(vm.versions, version)
		}
	}

	return nil
}

func (vm *VersionManager) GetVersions() []BackupVersion {
	return vm.versions
}

func (vm *VersionManager) GetVersion(id string) (*BackupVersion, error) {
	for _, ver := range vm.versions {
		if ver.ID == id {
			return &ver, nil
		}
	}
	return nil, fmt.Errorf("version not found: %s", id)
}

func (vm *VersionManager) GetLatestVersion() *BackupVersion {
	if len(vm.versions) == 0 {
		return nil
	}
	return &vm.versions[len(vm.versions)-1]
}
