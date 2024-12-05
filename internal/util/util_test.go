package util

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadConfigJSON tests the ReadConfig function with a JSON file.
func TestReadConfigJSON(t *testing.T) {
	configContent := `{
		"source_directory": "/tmp/source",
		"folders_to_backup": ["folder1", "folder2"],
		"target_directory": "/tmp/target",
		"deep_duplicate_check": true
	}`

	filePath := createTempFile(t, "config.json", configContent)
	defer os.Remove(filePath)

	config, err := ReadConfig(filePath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if config.SourceDirectory != "/tmp/source" {
		t.Errorf("Expected SourceDirectory '/tmp/source', got '%s'", config.SourceDirectory)
	}
	if len(config.FoldersToBackup) != 2 {
		t.Errorf("Expected 2 folders to backup, got %d", len(config.FoldersToBackup))
	}
	if config.TargetDirectory != "/tmp/target" {
		t.Errorf("Expected TargetDirectory '/tmp/target', got '%s'", config.TargetDirectory)
	}
	if !config.DeepDuplicateCheck {
		t.Errorf("Expected DeepDuplicateCheck to be true")
	}
}

// TestReadConfigYAML tests the ReadConfig function with a YAML file.
func TestReadConfigYAML(t *testing.T) {
	configContent := `
source_directory: /tmp/source
folders_to_backup:
  - folder1
  - folder2
target_directory: /tmp/target
deep_duplicate_check: true
`

	filePath := createTempFile(t, "config.yaml", configContent)
	defer os.Remove(filePath)

	config, err := ReadConfig(filePath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	if config.SourceDirectory != "/tmp/source" {
		t.Errorf("Expected SourceDirectory '/tmp/source', got '%s'", config.SourceDirectory)
	}
	if len(config.FoldersToBackup) != 2 {
		t.Errorf("Expected 2 folders to backup, got %d", len(config.FoldersToBackup))
	}
	if config.TargetDirectory != "/tmp/target" {
		t.Errorf("Expected TargetDirectory '/tmp/target', got '%s'", config.TargetDirectory)
	}
	if !config.DeepDuplicateCheck {
		t.Errorf("Expected DeepDuplicateCheck to be true")
	}
}

func TestValidate(t *testing.T) {
	// Create temporary source and target directories
	sourceDir := t.TempDir()
	targetDir := t.TempDir()

	config := &BackupConfig{
		SourceDirectory:   sourceDir,
		FoldersToBackup:   []string{"folder1", "folder2"},
		TargetDirectory:   targetDir,
		DeepDuplicateCheck: true,
	}

	// Case 1: Valid configuration
	if err := Validate(config); err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Case 2: Missing source directory
	os.RemoveAll(sourceDir) // Delete the source directory
	if err := Validate(config); err == nil || !strings.Contains(err.Error(), "source_directory") {
		t.Errorf("Expected validation error for missing source_directory, got: %v", err)
	}

	// Case 3: Empty target directory
	config.SourceDirectory = t.TempDir() // Recreate source directory
	config.TargetDirectory = ""
	if err := Validate(config); err == nil || !strings.Contains(err.Error(), "target_directory") {
		t.Errorf("Expected validation error for empty target_directory, got: %v", err)
	}

	// Case 4: Empty folders_to_backup
	config.TargetDirectory = t.TempDir() // Recreate target directory
	config.FoldersToBackup = []string{}
	if err := Validate(config); err == nil || !strings.Contains(err.Error(), "folders_to_backup") {
		t.Errorf("Expected validation error for empty folders_to_backup, got: %v", err)
	}
}

// TestPrintUsage tests the PrintUsage function.
func TestPrintUsage(t *testing.T) {
	config := &BackupConfig{
		SourceDirectory:   "/tmp/source",
		FoldersToBackup:   []string{"folder1", "folder2"},
		TargetDirectory:   "/tmp/target",
		DeepDuplicateCheck: true,
	}

	// Capture the output of PrintUsage
	output := captureOutput(func() {
		PrintUsage(config)
	})

	if !strings.Contains(output, "Copy from: /tmp/source/folder1 -> To: /tmp/target/folder1") {
		t.Errorf("Expected folder1 mapping in output, got: %s", output)
	}
	if !strings.Contains(output, "Copy from: /tmp/source/folder2 -> To: /tmp/target/folder2") {
		t.Errorf("Expected folder2 mapping in output, got: %s", output)
	}
}

//
// Helper Functions
//

// createTempFile creates a temporary file with the given content and returns the file path.
func createTempFile(t *testing.T, name string, content string) string {
	t.Helper()

	dir := os.TempDir()
	filePath := filepath.Join(dir, name)

	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	return filePath
}

// captureOutput captures the output of a function for testing.
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf strings.Builder
	_, _ = io.Copy(&buf, r)
	return buf.String()
}