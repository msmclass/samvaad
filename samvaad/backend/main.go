package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
)

//go:embed frontend_out
var frontendFS embed.FS

func main() {
	// Extract the embedded static Next.js frontend files
	subFS, err := fs.Sub(frontendFS, "frontend_out")
	if err != nil {
		fmt.Printf("failed to load embedded frontend files: %v\n", err)
		os.Exit(1)
	}

	// Serve the static files from the root path
	http.Handle("/", http.FileServer(http.FS(subFS)))

	// Port configuration (taken from environment variables)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("====================================================\n")
	fmt.Printf("🚀 SAMVAAD MEET SERVER\n")
	fmt.Printf("Sovereign Bharat SFU & Web App Platform\n")
	fmt.Printf("UI and backend unified on http://localhost:%s\n", port)
	fmt.Printf("====================================================\n")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("HTTP server exited with error: %v\n", err)
		os.Exit(1)
	}
}
