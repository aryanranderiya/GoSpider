package utils

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// downloadImage saves an image to the images folder within the domain directory
func DownloadImage(imageData []byte, urlStr string) {
	parsedURL, _ := url.Parse(urlStr)

	// Get domain
	domain := strings.TrimPrefix(parsedURL.Host, "www.")
	if domain == "" {
		domain = "unknown"
	}

	// Create domain/images folder
	outputDir := filepath.Join("output", domain, "images")
	os.MkdirAll(outputDir, 0755)

	// Simple filename: use the last part of URL or default
	filename := filepath.Base(parsedURL.Path)
	if filename == "/" || filename == "." || filename == "" {
		filename = "image.jpg"
	}

	filePath := filepath.Join(outputDir, filename)
	os.WriteFile(filePath, imageData, 0644)
	fmt.Printf("Downloaded: %s -> %s\n", urlStr, filePath)
}
