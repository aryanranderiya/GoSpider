package internal

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
)

func ConvertToMarkdown(input string, url string) string {

	markdown, err := htmltomarkdown.ConvertString(input,
		converter.WithDomain(url), // Convert relative links to absolute
	)
	if err != nil {
		log.Fatal(err)
	}

	return markdown
}

var (
	dirCache   = make(map[string]bool)
	dirCacheMu sync.RWMutex
)

// ensureDir creates directory only if it doesn't exist (cached)
func ensureDir(dir string) error {
	dirCacheMu.RLock()
	if dirCache[dir] {
		dirCacheMu.RUnlock()
		return nil
	}
	dirCacheMu.RUnlock()

	dirCacheMu.Lock()
	defer dirCacheMu.Unlock()

	// Double-check after acquiring write lock
	if dirCache[dir] {
		return nil
	}

	err := os.MkdirAll(dir, 0755)
	if err == nil {
		dirCache[dir] = true
	}
	return err
}

// SaveMarkdownToFile saves the markdown content to a file organized by domain
func SaveMarkdownToFile(markdown, urlStr string, verbose bool) {
	// Parse URL to get domain and path
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		if verbose {
			fmt.Println("Error parsing url to get domain and path", err)
		}
		return
	}

	// Get domain (remove www. if present)
	domain := strings.TrimPrefix(parsedURL.Host, "www.")
	if domain == "" {
		domain = "unknown"
	}

	// Create domain folder (cached)
	outputDir := filepath.Join("output", domain)
	if err := ensureDir(outputDir); err != nil {
		if verbose {
			fmt.Printf("Error creating directory %s: %v\n", outputDir, err)
		}
		return
	}

	// Get filename from URL path
	filename := filepath.Base(parsedURL.Path)
	if filename == "/" || filename == "." {
		filename = "index"
	}

	// Add .md extension if not present
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	filePath := filepath.Join(outputDir, filename)

	// Use buffered writing for better I/O performance
	file, err := os.Create(filePath)
	if err != nil {
		if verbose {
			fmt.Printf("Error creating file %s: %v\n", filePath, err)
		}
		return
	}
	defer file.Close()

	// Use buffered writer for faster I/O
	bufWriter := bufio.NewWriterSize(file, 262144) // 256kb buffer
	defer bufWriter.Flush()

	_, err = bufWriter.WriteString(markdown)
	if err != nil {
		if verbose {
			fmt.Printf("Error writing to file %s: %v\n", filePath, err)
		}
		return
	}

	if verbose {
		fmt.Printf("Saved: %s -> %s\n", urlStr, filePath)
	}
}
