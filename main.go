package main

import (
	"flag"
	"fmt"
	"gospider/internal"
	"gospider/utils"
	"sync"
	"time"
)

func main() {
	// Define command line flags
	startURL := flag.String("url", "", "Starting URL to crawl (required)")
	maxDomains := flag.Int("domains", 100, "Maximum number of domains to crawl (default 100)")
	numWorkers := flag.Int("workers", 5, "Number of concurrent workers (default 5)")
	useProxies := flag.Bool("proxies", false, "Use proxies from proxies.txt file")

	// Parse command line flags
	flag.Parse()

	// Validate required flags
	if *startURL == "" {
		fmt.Println("Error: -url flag is required")
		fmt.Println("Usage: go run main.go -url=https://example.com [-domains=3] [-workers=5]")
		flag.PrintDefaults()
		return
	}

	fmt.Println("\nwelcome to gospider by aryan randeriya\n")
	fmt.Printf("Starting URL: %s\n", *startURL)
	fmt.Printf("Max domains: %d\n", *maxDomains)
	fmt.Printf("Workers: %d\n", *numWorkers)
	fmt.Printf("Using proxies: %t\n", *useProxies)

	// Load proxies if requested
	if *useProxies {
		fmt.Println("Loading proxies...")
		utils.LoadProxies("proxies.txt")
	}

	// Initialize your custom queue - this stores URLs waiting to be processed
	queue := internal.NewQueue(*maxDomains)
	queue.Enqueue(*startURL) // Add the starting URL to begin crawling

	// Create a channel to communicate URLs between main thread and worker threads
	urlChannel := make(chan string, 200) // buffered channel with larger capacity
	var wg sync.WaitGroup                // WaitGroup tracks how many workers are currently processing URLs

	// Start multiple worker goroutines (run in background)
	fmt.Printf("Starting %d workers...\n", *numWorkers)
	for i := 0; i < *numWorkers; i++ {
		go internal.ProcessAllUrls(urlChannel, &wg, queue)
	}

	// Main loop: Move URLs from our queue to the channel for workers to process
	consecutiveEmptyChecks := 0
	for {
		// Try to get a URL from the queue
		url, successfullyPopped := queue.Dequeue()

		if !successfullyPopped {
			consecutiveEmptyChecks++

			// Give workers time to discover and add new URLs
			time.Sleep(100 * time.Millisecond)

			// If queue has been empty for multiple checks, wait for workers
			if consecutiveEmptyChecks >= 5 {
				// Wait for all active workers to complete
				wg.Wait()

				// Final check after workers are done
				url, successfullyPopped = queue.Dequeue()
				if !successfullyPopped {
					// Queue is still empty after workers finished
					if queue.IsCrawlingComplete() {
						fmt.Println("Crawling complete - domain limit reached and queue is empty.")
					} else {
						fmt.Println("All URLs processed. Shutting down...")
					}
					break
				}
				// Reset counter if we found a URL
				consecutiveEmptyChecks = 0
			} else {
				// Try to get another URL before the next iteration
				continue
			}
		} else {
			// Reset counter on successful dequeue
			consecutiveEmptyChecks = 0
		}

		// We got a URL from the queue successfully
		wg.Add(1)         // Tell WaitGroup we're starting a new task
		urlChannel <- url // Send the URL to worker for processing
	}

	// Close the channel to tell the worker goroutine to stop
	// This signals that no more URLs will be sent
	close(urlChannel)

	// Print final statistics
	fmt.Printf("\n=== Crawling Complete ===\n")
	fmt.Printf("Total unique URLs discovered: %d\n", queue.VisitedCount())
	fmt.Printf("URLs remaining in queue: %d\n", queue.Len())
}
