package util

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadConfig validates that the configuration file is correctly read and parsed.
func TestReadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
	{
		"source_directory": "/tmp/source",
		"target_directory": "/tmp/target",
		"folders_to_backup": [
			"folder1",
			"folder2",
			"folder3/subfolder"
		]
	}`
	tempFile := filepath.Join(os.TempDir(), "test_config.json")
	err := os.WriteFile(tempFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create temporary config file: %v", err)
	}
	defer os.Remove(tempFile)

	// Call ReadConfig
	config, err := ReadConfig(tempFile)
	if err != nil {
		t.Fatalf("ReadConfig failed: %v", err)
	}

	// Assert the values
	if config.SourceDirectory != "/tmp/source" {
		t.Errorf("Expected source_directory to be '/tmp/source', got '%s'", config.SourceDirectory)
	}
	if config.TargetDirectory != "/tmp/target" {
		t.Errorf("Expected target_directory to be '/tmp/target', got '%s'", config.TargetDirectory)
	}
	if len(config.FoldersToBackup) != 3 {
		t.Errorf("Expected 3 folders to backup, got %d", len(config.FoldersToBackup))
	}
}

// TestValidateFolders tests the validation logic for source, target, and folders_to_backup.
func TestValidateFolders(t *testing.T) {
	// Create temporary directories
	sourceDir := filepath.Join(os.TempDir(), "test_source")
	targetDir := filepath.Join(os.TempDir(), "test_target")
	folder1 := filepath.Join(sourceDir, "folder1")
	folder2 := filepath.Join(sourceDir, "folder2")
	subfolder := filepath.Join(sourceDir, "folder3", "subfolder")

	os.MkdirAll(folder1, 0755)
	os.MkdirAll(folder2, 0755)
	os.MkdirAll(subfolder, 0755)
	defer os.RemoveAll(sourceDir)
	defer os.RemoveAll(targetDir)

	// Test with valid directories
	err := ValidateFolders(sourceDir, targetDir, []string{"folder1", "folder2", "folder3/subfolder"})
	if err != nil {
		t.Fatalf("ValidateFolders failed: %v", err)
	}

	// Check if target directory was created
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		t.Errorf("Target directory was not created: %s", targetDir)
	}

	// Test with missing folder
	err = ValidateFolders(sourceDir, targetDir, []string{"nonexistent"})
	if err != nil {
		t.Errorf("ValidateFolders returned an error for missing folder: %v", err)
	}
}

// TestPrintUsage verifies that the PrintUsage function generates the correct output.
func TestPrintUsage(t *testing.T) {
	// Mock a configuration
	config := Config{
		SourceDirectory: "/mock/source",
		TargetDirectory: "/mock/target",
		FoldersToBackup: []string{"folder1", "folder2", "folder3/subfolder"},
	}

	// Capture the output
	output := &strings.Builder{}
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintUsage(config)

	w.Close()
	os.Stdout = oldStdout
	r.ReadString(0)

	// Assert output
	expectedOutput := `
Backup Plan:
Copy from: /mock/source/folder1 -> To: /mock/target/folder1
Copy from: /mock/source/folder2 -> To: /mock/target/folder2
Copy from: /mock/source/folder3/subfolder -> To: /mock/target/folder3/subfolder

----------------
All folders validated successfully. Ready to start the backup process.
`

	if output.String() != expectedOutput {
		t.Errorf("Expected:\n%s\nGot:\n%s", expectedOutput, output.String())
	}
}
