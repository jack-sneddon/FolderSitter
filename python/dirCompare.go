package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ComparisonResult struct {
	Added    []string
	Removed  []string
	Changed  []string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <config-file>")
		return
	}

	configFile := os.Args[1]

	// Read configuration
	originDir, destinationDir, err := readConfig(configFile)
	if err != nil {
		fmt.Println("Error reading config:", err)
		return
	}

	// Validate directories
	if !dirExists(originDir) {
		fmt.Printf("Error: Origin directory does not exist: %s\n", originDir)
		return
	}
	if !dirExists(destinationDir) {
		fmt.Printf("Error: Destination directory does not exist: %s\n", destinationDir)
		return
	}

	fmt.Printf("Comparing directories:\nOrigin: %s\nDestination: %s\n", originDir, destinationDir)

	startTime := time.Now()

	// Calculate folder sizes
	originSize, err := calculateFolderSize(originDir)
	if err != nil {
		fmt.Println("Error calculating origin folder size:", err)
		return
	}

	destinationSize, err := calculateFolderSize(destinationDir)
	if err != nil {
		fmt.Println("Error calculating destination folder size:", err)
		return
	}

	// Compute checksums
	originChecksums, err := computeFolderChecksums(originDir)
	if err != nil {
		fmt.Println("Error computing checksums for origin:", err)
		return
	}

	destinationChecksums, err := computeFolderChecksums(destinationDir)
	if err != nil {
		fmt.Println("Error computing checksums for destination:", err)
		return
	}

	// Compare folders
	comparison := compareFolders(originChecksums, destinationChecksums)

	// Calculate elapsed time
	elapsed := time.Since(startTime)

	// Write results to output file
	outputFile := fmt.Sprintf("out/%s.out", filepath.Base(originDir))
	err = writeResults(outputFile, originDir, destinationDir, originSize, destinationSize, elapsed, comparison)
	if err != nil {
		fmt.Println("Error writing results:", err)
		return
	}

	fmt.Printf("Comparison complete. Results written to %s\n", outputFile)
}

func readConfig(configFile string) (string, string, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return "", "", err
	}
	defer file.Close()

	var originDir, destinationDir string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "original=") {
			originDir = strings.Trim(strings.SplitN(line, "=", 2)[1], `"`)
		} else if strings.HasPrefix(line, "destination=") {
			destinationDir = strings.Trim(strings.SplitN(line, "=", 2)[1], `"`)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", err
	}

	if originDir == "" || destinationDir == "" {
		return "", "", errors.New("both original and destination directories must be specified")
	}

	return originDir, destinationDir, nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func calculateFolderSize(root string) (int64, error) {
	var totalSize int64
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})
	return totalSize, err
}

func computeFolderChecksums(root string) (map[string]string, error) {
	checksums := make(map[string]string)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(root, path)
			hash, err := computeFileChecksum(path)
			if err != nil {
				return err
			}
			checksums[relPath] = hash
		}
		return nil
	})
	return checksums, err
}

func computeFileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func compareFolders(origin, destination map[string]string) ComparisonResult {
	originFiles := make(map[string]bool)
	for file := range origin {
		originFiles[file] = true
	}

	destinationFiles := make(map[string]bool)
	for file := range destination {
		destinationFiles[file] = true
	}

	added := []string{}
	removed := []string{}
	changed := []string{}

	for file := range origin {
		if !destinationFiles[file] {
			removed = append(removed, file)
		} else if origin[file] != destination[file] {
			changed = append(changed, file)
		}
	}

	for file := range destination {
		if !originFiles[file] {
			added = append(added, file)
		}
	}

	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)

	return ComparisonResult{Added: added, Removed: removed, Changed: changed}
}

func writeResults(outputFile, originDir, destinationDir string, originSize, destinationSize int64, elapsed time.Duration, comparison ComparisonResult) error {
	err := os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	_, _ = fmt.Fprintf(writer, "Run at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	_, _ = fmt.Fprintf(writer, "Time to complete - %s\n\n", elapsed.Round(time.Second))
	_, _ = fmt.Fprintf(writer, "Size\n--------\n")
	_, _ = fmt.Fprintf(writer, "Origin Folder: %d MB\n", originSize/1024/1024)
	_, _ = fmt.Fprintf(writer, "Destination Folder: %d MB\n\n", destinationSize/1024/1024)
	_, _ = fmt.Fprintf(writer, "Comparison Results\n--------------------\n")
	_, _ = fmt.Fprintf(writer, "Added Files: %v\n", comparison.Added)
	_, _ = fmt.Fprintf(writer, "Removed Files: %v\n", comparison.Removed)
	_, _ = fmt.Fprintf(writer, "Changed Files: %v\n", comparison.Changed)

	return nil
}