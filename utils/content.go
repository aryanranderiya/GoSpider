package utils

import "strings"

// isHTML checks if the content type is HTML
func IsHTML(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "text/html")
}

// isImage checks if the content type is an image
func IsImage(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "image/")
}
