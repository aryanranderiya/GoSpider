package main

import (
	"fmt"
	"gospider/internal"
	"sync"
)

func main() {

	fmt.Println("welcome to gospider by aryan randeriya")

	var startURL string
	fmt.Println("Enter the starting url:")
	fmt.Scanln(&startURL)

	// Initialize your custom queue - this stores URLs waiting to be processed
	queue := internal.NewQueue()
	queue.Enqueue(startURL) // Add the starting URL to begin crawling

	// Create a channel to communicate URLs between main thread and worker threads
	urlChannel := make(chan string, 100) // buffered channel with capacity 100
	var wg sync.WaitGroup                // WaitGroup tracks how many workers are currently processing URLs

	// Start one worker goroutine (runs in background)
	// This worker will read URLs from urlChannel and process them
	go internal.ProcessAllUrls(urlChannel, &wg, queue)

	// Main loop: Move URLs from our queue to the channel for workers to process
	for {
		// Try to get a URL from the queue
		url, successfullyPopped := queue.Dequeue()

		if !successfullyPopped {
			// Queue is empty right now, but workers might still be running  and could add more URLs to the queue

			// Wait for all current workers to finish their tasks
			if wg.Wait(); true {
				// Now that all workers are done, check if they added any new URLs
				url, successfullyPopped = queue.Dequeue()
				if !successfullyPopped {
					// Queue is still empty and no workers are running
					// This means we've processed everything - time to exit
					break
				}
				// Found a new URL! Tell WaitGroup we're starting another task
				wg.Add(1)
				urlChannel <- url // Send URL to worker
				continue          // Go back to start of loop
			}
			break // This should never happen due to the 'true' condition above
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
