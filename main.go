package main

import (
	"fmt"
	"gospider/internal"
	"sync"
	"time"
)

func main() {

	fmt.Println("welcome to gospider by aryan randeriya")

	var startURL string
	fmt.Println("Enter the starting url:")
	fmt.Scanln(&startURL)

	var maxDomains int
	fmt.Println("Enter maximum number of domains to crawl (e.g., 3):")
	fmt.Scanln(&maxDomains)

	var numWorkers int
	fmt.Println("Enter number of workers (e.g., 5 for 5x speed):")
	fmt.Scanln(&numWorkers)

	// Initialize your custom queue - this stores URLs waiting to be processed
	queue := internal.NewQueue(maxDomains)
	queue.Enqueue(startURL) // Add the starting URL to begin crawling

	// Create a channel to communicate URLs between main thread and worker threads
	urlChannel := make(chan string, 200) // buffered channel with larger capacity
	var wg sync.WaitGroup                // WaitGroup tracks how many workers are currently processing URLs

	// Start multiple worker goroutines (run in background)
	fmt.Printf("Starting %d workers...\n", numWorkers)
	for i := 0; i < numWorkers; i++ {
		go internal.ProcessAllUrls(urlChannel, &wg, queue)
	}

	// Main loop: Move URLs from our queue to the channel for workers to process
	for {
		// Try to get a URL from the queue
		url, successfullyPopped := queue.Dequeue()

		if !successfullyPopped {
			// Queue is empty - check if we've hit the domain limit
			if queue.IsCrawlingComplete() {
				fmt.Println("Domain limit reached and no more URLs to process. Waiting for workers to finish...")
				wg.Wait()
				// Double-check after workers finish
				if queue.IsCrawlingComplete() {
					fmt.Println("Crawling complete - domain limit reached.")
					break
				}
			}

			// Queue is empty - give workers some time to finish and add more URLs
			time.Sleep(100 * time.Millisecond)

			// Try again after a short wait
			url, successfullyPopped = queue.Dequeue()
			if !successfullyPopped {
				// Still empty - wait for all workers to finish
				wg.Wait()

				// Final check after all workers are done
				url, successfullyPopped = queue.Dequeue()
				if !successfullyPopped {
					// Check if crawling is complete or just no more URLs
					if queue.IsCrawlingComplete() {
						fmt.Println("Crawling complete - domain limit reached.")
					} else {
						fmt.Println("All URLs processed. Shutting down...")
					}
					break
				}
			}
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
