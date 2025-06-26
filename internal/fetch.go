package internal

import (
	"fmt"
	"gospider/utils"
	"io"
	"net/http"
	"sync"
)

func Fetch(url string, wg *sync.WaitGroup, queue *Queue, downloadImages bool, saveFiles bool, verbose bool) {
	// Use shared HTTP client with connection pooling
	client := GetHTTPClient(verbose)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

	response, err := client.Do(req)

	if err != nil {
		if verbose {
			fmt.Println("Error while trying to fetch url: ", url, err)
		}
		return
	}

	// always close the request body to prevent resource leakage
	defer response.Body.Close()

	// Get content type to determine how to handle the content
	contentType := response.Header.Get("Content-Type")

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		if verbose {
			fmt.Println("Error reading body:", err)
		}
		return
	}

	// Check if it's an image
	if utils.IsImage(contentType) {
		if downloadImages {
			go utils.DownloadImage(body, url, verbose)
		}
		return
	}

	// Check if it's processable HTML/text content
	if !utils.IsHTML(contentType) {
		if verbose {
			fmt.Printf("Skipping non-HTML content: %s (Content-Type: %s)\n", url, contentType)
		}
		return
	}

	// Process HTML content
	markdown := ConvertToMarkdown(string(body), url)

	// Save to file only if saveFiles flag is enabled
	if saveFiles {
		SaveMarkdownToFile(markdown, url, verbose)
	}

	urls := utils.ExtractURLs(string(body))
	urls_md := utils.ExtractURLs(markdown)

	// Combine URLs from both sources and remove duplicates
	urlSet := make(map[string]bool)
	for _, u := range urls {
		urlSet[u] = true
	}
	for _, u := range urls_md {
		urlSet[u] = true
	}

	// Iterate over all unique urls and add to the queue for processing
	for url := range urlSet {
		if verbose {
			fmt.Println("Found URL:", url)
		}
		queue.Enqueue(url)
	}

	// Mark this URL as successfully completed
	queue.MarkCompleted()
}
