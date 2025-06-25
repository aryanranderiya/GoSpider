package utils

import (
	"regexp"
)

func ExtractURLs(text string) []string {
	// Basic regex to match HTTP/HTTPS URLs
	re := regexp.MustCompile(`https?://[^\s"'>]+`)
	return re.FindAllString(text, -1)
}
