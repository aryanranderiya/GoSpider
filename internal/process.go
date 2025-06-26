package internal

import (
	"fmt"
	"sync"
)

func ProcessAllUrls(urlChan <-chan string, wg *sync.WaitGroup, queue *Queue, downloadImages bool, saveFiles bool, verbose bool) {
	// This function runs in a separate goroutine (worker thread)
	// It reads URLs from the channel and processes them one by one
	for url := range urlChan {
		if verbose {
			fmt.Println("Processing:", url)
		}
		// Download and parse the HTML content
		Fetch(url, wg, queue, downloadImages, saveFiles, verbose)

		// Tell the WaitGroup that this worker has finished processing this URL
		wg.Done()
	}
}
