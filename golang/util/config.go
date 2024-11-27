package util

import (
	"encoding/json"
	"fmt"
	"os"
)

// BackupConfig represents the configuration structure for the backup process.
type BackupConfig struct {
	SourceDirectory    string   `json:"source_directory"`
	TargetDirectory    string   `json:"target_directory"`
	FoldersToBackup    []string `json:"folders_to_backup"`
	DeepDuplicateCheck bool     `json:"deep_duplicate_check"`
}

// ReadConfig reads the JSON configuration file and returns a BackupConfig struct.
func ReadConfig(filePath string) (BackupConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return BackupConfig{}, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	var config BackupConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return BackupConfig{}, fmt.Errorf("error decoding config file: %v", err)
	}

	fmt.Printf("Loaded configuration:\n%+v\n", config)
	return config, nil
}
