package internal

import (
	"fmt"
	"log"
	"net/url"
	"os"
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

// SaveMarkdownToFile saves the markdown content to a file organized by domain
func SaveMarkdownToFile(markdown, urlStr string) {
	// Parse URL to get domain and path
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("Error parsing url to get domain and path", err)
	}

	// Get domain (remove www. if present)
	domain := strings.TrimPrefix(parsedURL.Host, "www.")
	if domain == "" {
		domain = "unknown"
	}

	// Create domain folder
	outputDir := filepath.Join("output", domain)
	os.MkdirAll(outputDir, 0755)

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

	// Write file
	os.WriteFile(filePath, []byte(markdown), 0644)

	fmt.Printf("Saved: %s -> %s\n", urlStr, filePath)
}
