package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func ExtractURLs(text string) []string {
	// Basic regex to match HTTP/HTTPS URLs
	re := regexp.MustCompile(`https?://[^\s"'>]+`)
	return re.FindAllString(text, -1)
}

func ExtractDomain(urlStr string, verbose bool) (string, bool) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil || parsedURL.Host == "" {
		if verbose {
			fmt.Printf("Skipping invalid URL: %s\n", urlStr)
		}
		return "", false
	}

	domain := strings.TrimPrefix(parsedURL.Host, "www.")
	if domain == "" {
		if verbose {
			fmt.Printf("Skipping URL with empty domain: %s\n", urlStr)
		}
		return "", false
	}

	return domain, true
}
