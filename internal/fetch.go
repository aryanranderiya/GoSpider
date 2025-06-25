package internal

import (
	"fmt"
	"gospider/utils"
	"io"
	"net/http"
	"sync"
	"time"
)

func Fetch(url string, wg *sync.WaitGroup, queue *Queue) {
	wg.Add(1)

	client := &http.Client{
		Timeout: 10 * time.Second, // 10 second timeout to prevent hanging
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "GoSpider/1.0")

	response, err := client.Do(req)

	if err != nil {
		fmt.Println("Error while trying to fetch url: ", url, err)
		return
	}

	// always close the request body to prevent resource leakage
	defer response.Body.Close()

	// Get content type to determine how to handle the content
	contentType := response.Header.Get("Content-Type")

	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading body:", err)
		return
	}

	// Check if it's an image
	if utils.IsImage(contentType) {
		utils.DownloadImage(body, url)
		return
	}

	// Check if it's processable HTML/text content
	if !utils.IsHTML(contentType) {
		fmt.Printf("Skipping non-HTML content: %s (Content-Type: %s)\n", url, contentType)
		return
	}

	// Process HTML content
	markdown := ConvertToMarkdown(string(body), url)
	SaveMarkdownToFile(markdown, url)

	urls := utils.ExtractURLs(string(body))

	// Iterate over all urls and add to the queue for processing
	for _, url := range urls {
		fmt.Println("Found URL:", url)
		queue.Enqueue(url)
	}

	wg.Done()
}
