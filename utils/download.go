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

// DownloadImage saves an image to the images folder within the domain directory
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

	// Create domain/images folder (cached)
	outputDir := filepath.Join("output", domain, "images")
	if err := ensureImageDir(outputDir); err != nil {
		if verbose {
			fmt.Printf("Error creating image directory %s: %v\n", outputDir, err)
		}
		return
	}

	// Simple filename: use the last part of URL or default
	filename := filepath.Base(parsedURL.Path)
	if filename == "/" || filename == "." || filename == "" {
		filename = "image.jpg"
	}

	filePath := filepath.Join(outputDir, filename)

	// Use buffered writing for better I/O performance
	file, err := os.Create(filePath)
	if err != nil {
		if verbose {
			fmt.Printf("Error creating image file %s: %v\n", filePath, err)
		}
		return
	}
	defer file.Close()

	// Use buffered writer for faster I/O
	bufWriter := bufio.NewWriterSize(file, 65536) // 64KB buffer for images
	defer bufWriter.Flush()

	_, err = bufWriter.Write(imageData)
	if err != nil {
		if verbose {
			fmt.Printf("Error writing image file %s: %v\n", filePath, err)
		}
		return
	}

	if verbose {
		fmt.Printf("Downloaded: %s -> %s\n", urlStr, filePath)
	}
}
