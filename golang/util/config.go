package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// BackupConfig represents the configuration for the backup process.
type BackupConfig struct {
	SourceDirectory   string   `json:"source_directory" yaml:"source_directory"`
	FoldersToBackup   []string `json:"folders_to_backup" yaml:"folders_to_backup"`
	TargetDirectory   string   `json:"target_directory" yaml:"target_directory"`
	DeepDuplicateCheck bool     `json:"deep_duplicate_check" yaml:"deep_duplicate_check"`
}

func ReadConfig(filePath string) (*BackupConfig, error) {
	ext := filepath.Ext(filePath)

	config := &BackupConfig{}
	var err error

	switch ext {
	case ".json":
		err = readJSONConfig(filePath, config)
	case ".yaml", ".yml":
		err = readYAMLConfig(filePath, config)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return nil, err
	}
	return config, nil
}

// readJSONConfig reads configuration data from a JSON file.
func readJSONConfig(filePath string, config *BackupConfig) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening JSON file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("error decoding JSON file: %v", err)
	}
	return nil
}

// readYAMLConfig reads configuration data from a YAML file.
func readYAMLConfig(filePath string, config *BackupConfig) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening YAML file: %v", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return fmt.Errorf("error decoding YAML file: %v", err)
	}
	return nil
}