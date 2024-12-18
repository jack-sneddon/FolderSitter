echo "\n\n******* dry-run ******\n\n"
go run cmd/main.go -config configs/test_config.yaml --dry-run
rm -rf /Users/jack/tmp/2024-11-27-backyaml

echo "\n\n******* copy *****\n\n"
go run cmd/main.go -config configs/test_config.yaml
du -h /Users/jack/tmp/2024-11-27-backyaml
