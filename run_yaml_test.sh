echo "\n\n******* dry-run ******\n\n"
go run cmd/main.go -config configs/test_config.yaml --dry-run
rm -rf /Users/jack/tmp/2024-11-27-backyaml

echo "\n\n******* copy *****\n\n"
go run cmd/main.go -config configs/test_config.yaml
du -h /Users/jack/tmp/2024-11-27-backyaml

echo "\n\n******* list versions *****\n\n"
# List all versions
go run cmd/main.go -config configs/test_config.yaml --list-versions

# Show specific version
echo "\n\n******* show specific versions *****\n\n"
go run cmd/main.go -config configs/test_config.yaml --show-version 20231217-212653

# Show latest version
echo "\n\n******* show latest versions *****\n\n"
go run cmd/main.go -config configs/test_config.yaml --latest-version
