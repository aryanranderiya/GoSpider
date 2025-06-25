package utils

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	imageDirCache   = make(map[string]bool)
	imageDirCacheMu sync.RWMutex
)

// ensureImageDir creates directory only if it doesn't exist (cached)
func ensureImageDir(dir string) error {
	imageDirCacheMu.RLock()
	if imageDirCache[dir] {
		imageDirCacheMu.RUnlock()
		return nil
	}
	imageDirCacheMu.RUnlock()

	imageDirCacheMu.Lock()
	defer imageDirCacheMu.Unlock()
	
	// Double-check after acquiring write lock
	if imageDirCache[dir] {
		return nil
	}
	
	err := os.MkdirAll(dir, 0755)
	if err == nil {
		imageDirCache[dir] = true
	}
	return err
}

// DownloadImage saves an image to the images folder within the domain directory using high-speed writer
func DownloadImage(imageData []byte, urlStr string, verbose bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		if verbose {
			fmt.Printf("Error parsing image URL %s: %v\n", urlStr, err)
		}
		return
	}

	// Get domain
	domain := strings.TrimPrefix(parsedURL.Host, "www.")
	if domain == "" {
		domain = "unknown"
	}

	// Simple filename: use the last part of URL or default
	filename := filepath.Base(parsedURL.Path)
	if filename == "/" || filename == "." || filename == "" {
		filename = "image.jpg"
	}

	// Build full file path
	outputDir := filepath.Join("output", domain, "images")
	filePath := filepath.Join(outputDir, filename)

	// Use high-speed file writer for maximum throughput
	// Note: Need to import internal package or move this function
	// For now, fall back to direct writing
	ensureImageDir(outputDir)
	file, err := os.Create(filePath)
	if err != nil {
		if verbose {
			fmt.Printf("Error creating image file %s: %v\n", filePath, err)
		}
		return
	}
	defer file.Close()

	bufWriter := bufio.NewWriterSize(file, 1048576) // 1MB buffer
	defer bufWriter.Flush()
	bufWriter.Write(imageData)

	if verbose {
		fmt.Printf("Downloaded: %s -> %s\n", urlStr, filePath)
	}
}
