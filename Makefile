# Apple Music Downloader Makefile

.PHONY: help build run-server run-cli clean install-deps

# Default target
help:
	@echo "Apple Music Downloader - Available commands:"
	@echo "  build        - Build the application"
	@echo "  run-server   - Run the web server"
	@echo "  run-cli      - Run the command line interface"
	@echo "  clean        - Clean build artifacts"
	@echo "  install-deps - Install Go dependencies"
	@echo "  setup        - Setup configuration file"

# Install Go dependencies
install-deps:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy

# Setup configuration
setup:
	@echo "Setting up configuration..."
	@if [ ! -f config.yaml ]; then \
		cp config.yaml.example config.yaml; \
		echo "Created config.yaml from example. Please edit it with your tokens."; \
	else \
		echo "config.yaml already exists."; \
	fi

# Build the application
build: install-deps
	@echo "Building Apple Music Downloader..."
	go build -o apple-music-downloader main.go
	go build -o apple-music-downloader-server web_server.go server.go main.go

# Run web server
run-server: setup
	@echo "Starting Apple Music Downloader Web Server..."
	@echo "Open your browser and go to: http://localhost:8080"
	go run web_server.go server.go main.go -port 8080

# Run command line interface
run-cli: setup
	@echo "Starting Apple Music Downloader CLI..."
	go run main.go

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f apple-music-downloader
	rm -f apple-music-downloader-server

# Build for different platforms
build-windows: install-deps
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 go build -o apple-music-downloader-windows.exe web_server.go server.go main.go

build-linux: install-deps
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 go build -o apple-music-downloader-linux web_server.go server.go main.go

build-mac: install-deps
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 go build -o apple-music-downloader-mac web_server.go server.go main.go

# Build all platforms
build-all: build-windows build-linux build-mac
	@echo "Built for all platforms"

# Development
dev: install-deps
	@echo "Starting development server with auto-reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Installing air for auto-reload..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# Test
test:
	@echo "Running tests..."
	go test ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi 