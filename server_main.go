package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Define command line flags
	port := flag.String("port", "8080", "Port to run the server on")
	flag.Parse()

	// Create and start server
	server := NewServer(*port)
	
	fmt.Println("🍎 Apple Music Downloader Web Server")
	fmt.Println("=====================================")
	fmt.Printf("Server will start on port: %s\n", *port)
	fmt.Println("Make sure you have configured your config.yaml file properly.")
	fmt.Println("Required tokens:")
	fmt.Println("- media-user-token: For downloading AAC-LC, lyrics, and music videos")
	fmt.Println("- authorization-token: Usually auto-obtained, but can be set manually")
	fmt.Println("- storefront: Your Apple Music storefront (e.g., 'us', 'jp', 'ca')")
	fmt.Println()

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 