package internal

import (
	"fmt"
	"gospider/utils"
	"io"
	"net/http"
	"sync"
)

func Fetch(url string, wg *sync.WaitGroup, queue *Queue, verbose bool) {
	// Create HTTP client with tested proxy fallback
	client := utils.CreateHTTPClientWithTestedProxy(verbose)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "GoSpider/1.0")

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
		utils.DownloadImage(body, url, verbose)
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
	SaveMarkdownToFile(markdown, url, verbose)

	urls := utils.ExtractURLs(markdown)

	// Iterate over all urls and add to the queue for processing
	for _, url := range urls {
		if verbose {
			fmt.Println("Found URL:", url)
		}
		queue.Enqueue(url)
	}
}
