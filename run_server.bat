@echo off
echo ========================================
echo Apple Music Downloader Web Server
echo ========================================
echo.

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/
    pause
    exit /b 1
)

REM Check if config.yaml exists
if not exist "config.yaml" (
    echo Error: config.yaml not found
    echo Please copy config.yaml.example to config.yaml and configure it
    pause
    exit /b 1
)

echo Starting server on port 8080...
echo Open your browser and go to: http://localhost:8080
echo.
echo Press Ctrl+C to stop the server
echo.

REM Run the server
go run web_server.go server.go main.go -port 8080

pause 