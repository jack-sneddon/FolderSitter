#!/bin/bash

# Clean build script for FolderSitter

# Step 1: Clean up old build artifacts
echo "Cleaning old build artifacts..."
rm -f foldersitter
rm -rf out/

# Step 2: Format Go source files
echo "Formatting Go source code..."
go fmt ./...

# Step 3 Clean cache
go clean -modcache

# Step 4: Tidy up dependencies
echo "Tidying up dependencies..."
go mod tidy

# Step 5: Verify module integrity
echo "Verifying module integrity..."
go mod verify

# Step 6: Clean Go cache (optional)
echo "Cleaning Go module cache..."
go clean -cache

# Step 7: Build the project
echo "Building the project..."
# go build -o foldersitter ./cmd/main.go
go build ./...

# Step 8: Run the binary to test
# ./foldersitter -config configs/run_test_config.yaml
