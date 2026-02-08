.PHONY: build run test test-v clean help

# Build the server binary
build:
	go build -o bin/server ./cmd/server

# Run the server
run:
	go run ./cmd/server

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Run tests with coverage
test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Show help
help:
	@echo "Available targets:"
	@echo "  build      - Build the server binary"
	@echo "  run        - Run the server"
	@echo "  test       - Run all tests"
	@echo "  test-v     - Run tests with verbose output"
	@echo "  test-cover - Run tests with coverage report"
	@echo "  clean      - Remove build artifacts"
