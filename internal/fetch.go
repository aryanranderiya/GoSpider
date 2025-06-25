package internal

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func ProcessAllUrls(urlChan <-chan string, wg *sync.WaitGroup, queue *Queue) {
	// This function runs in a separate goroutine (worker thread)
	// It reads URLs from the channel and processes them one by one
	for url := range urlChan {
		fmt.Println("Processing:", url)

		// Call the Fetch function to simulate fetching the web page
		// This represents downloading and parsing the HTML content
		Fetch(url, wg)

		// TODO:
		// 1. Fetch the web page content (done above)
		// 2. Parse HTML to find more links
		// 3. Add discovered links to the queue for further crawling

		// Simulate random behavior: sometimes find new URLs, sometimes don't
		// This makes testing more realistic
		if rand.Float32() < 0.9 { // 70% chance to find a new URL
			newURL := fmt.Sprintf("discovered-url-%d", rand.Intn(1000))
			fmt.Printf("  → Found new URL: %s\n", newURL)
			queue.Enqueue(newURL)
		} else {
			fmt.Println("  → No new URLs found on this page")
		}

		// Tell the WaitGroup that this worker has finished processing this URL
		wg.Done()
	}
}

func Fetch(url string, wg *sync.WaitGroup) {
	wg.Add(1)
	fmt.Println("doing some work")
	time.Sleep(1 * time.Second) // simulate work
	wg.Done()
}
