#!/bin/bash

echo "========================================"
echo "Apple Music Downloader Web Server"
echo "========================================"
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed or not in PATH"
    echo "Please install Go from https://golang.org/"
    exit 1
fi

# Check if config.yaml exists
if [ ! -f "config.yaml" ]; then
    echo "Error: config.yaml not found"
    echo "Please copy config.yaml.example to config.yaml and configure it"
    exit 1
fi

echo "Starting server on port 8080..."
echo "Open your browser and go to: http://localhost:8080"
echo
echo "Press Ctrl+C to stop the server"
echo

# Run the server
go run web_server.go server.go main.go -port 8080 