package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func ExtractURLs(text string) []string {
	// Improved regex to match HTTP/HTTPS URLs without trailing punctuation
	re := regexp.MustCompile(`https?://[^\s"'>\)]+`)
	matches := re.FindAllString(text, -1)
	
	// Clean up URLs by removing trailing punctuation
	var cleanURLs []string
	for _, match := range matches {
		// Remove trailing punctuation like ), ], }, ., ,, ;, :, !, ?
		cleaned := strings.TrimRight(match, ".,;:!?)]}")
		if cleaned != "" {
			cleanURLs = append(cleanURLs, cleaned)
		}
	}
	
	return cleanURLs
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
