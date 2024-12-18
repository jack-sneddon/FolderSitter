echo "\n******* dry-run ******\n"
go run cmd/main.go -config configs/test_config.yaml --dry-run
rm -rf /Users/jack/tmp/2024-11-27-backyaml

echo "\n******* copy *****\n"
go run cmd/main.go -config configs/test_config.yaml
du -h /Users/jack/tmp/2024-11-27-backyaml

echo "\n******* copy2 *****\n"
rm -rf /Users/jack/tmp/2024-11-27-backyaml/Packers
go run cmd/main.go -config configs/test_config.yaml

echo "\n******* list versions *****\n"
# List all versions
go run cmd/main.go -config configs/test_config.yaml --list-versions

# Show latest version
echo "\n******* show latest versions *****\n"
go run cmd/main.go -config configs/test_config.yaml --latest-version
