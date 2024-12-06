package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// BackupConfig represents the configuration for the backup process.
type BackupConfig struct {
	SourceDirectory    string   `json:"source_directory" yaml:"source_directory"`
	FoldersToBackup    []string `json:"folders_to_backup" yaml:"folders_to_backup"`
	TargetDirectory    string   `json:"target_directory" yaml:"target_directory"`
	DeepDuplicateCheck bool     `json:"deep_duplicate_check" yaml:"deep_duplicate_check"`
}

// Load reads the configuration file (JSON or YAML) and returns a BackupConfig instance.
func Load(filePath string) (*BackupConfig, error) {
	ext := filepath.Ext(filePath)

	config := &BackupConfig{}
	var err error

	switch ext {
	case ".json":
		err = readJSON(filePath, config)
	case ".yaml", ".yml":
		err = readYAML(filePath, config)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return config, nil
}

// readJSON reads a JSON configuration file.
func readJSON(filePath string, config *BackupConfig) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open JSON file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("failed to decode JSON file: %w", err)
	}

	return nil
}

// readYAML reads a YAML configuration file.
func readYAML(filePath string, config *BackupConfig) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open YAML file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("failed to decode YAML file: %w", err)
	}

	return nil
}

func Validate(cfg *BackupConfig) error {
	if cfg.SourceDirectory == "" {
		return fmt.Errorf("source_directory is empty")
	}
	if cfg.TargetDirectory == "" {
		return fmt.Errorf("target_directory is empty")
	}
	if len(cfg.FoldersToBackup) == 0 {
		return fmt.Errorf("folders_to_backup is empty")
	}
	return nil
}
