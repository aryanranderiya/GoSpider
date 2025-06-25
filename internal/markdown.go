package internal

import (
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"

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

// SaveMarkdownToFile saves the markdown content to a file organized by domain using high-speed writer
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

	// Get filename from URL path
	filename := filepath.Base(parsedURL.Path)
	if filename == "/" || filename == "." {
		filename = "index"
	}

	// Add .md extension if not present
	if !strings.HasSuffix(filename, ".md") {
		filename += ".md"
	}

	// Build full file path
	outputDir := filepath.Join("output", domain)
	filePath := filepath.Join(outputDir, filename)

	// Use high-speed file writer for maximum throughput
	fileWriter := GetFileWriter()
	fileWriter.WriteFile(filePath, []byte(markdown), verbose)
}
