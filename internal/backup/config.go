package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Options struct {
	Verbose  bool
	Quiet    bool
	LogLevel string
}

type Config struct {
	SourceDirectory    string        `json:"source_directory" yaml:"source_directory"`
	FoldersToBackup    []string      `json:"folders_to_backup" yaml:"folders_to_backup"`
	TargetDirectory    string        `json:"target_directory" yaml:"target_directory"`
	DeepDuplicateCheck bool          `json:"deep_duplicate_check" yaml:"deep_duplicate_check"`
	Concurrency        int           `json:"concurrency" yaml:"concurrency"`
	BufferSize         int           `json:"buffer_size" yaml:"buffer_size"`
	RetryAttempts      int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay         time.Duration `json:"retry_delay" yaml:"retry_delay"`
	ExcludePatterns    []string      `json:"exclude_patterns" yaml:"exclude_patterns"`
	ChecksumAlgorithm  string        `json:"checksum_algorithm" yaml:"checksum_algorithm"`
	Options            *Options
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, newBackupError("ReadConfig", path, err)
	}

	config := &Config{
		Concurrency:       4,
		BufferSize:        32 * 1024,
		RetryAttempts:     3,
		RetryDelay:        time.Second,
		ChecksumAlgorithm: "sha256",
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, config)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, config)
	default:
		return nil, newBackupError("LoadConfig", path, fmt.Errorf("unsupported format: %s", ext))
	}

	if err != nil {
		return nil, newBackupError("ParseConfig", path, err)
	}

	return config, nil
}
