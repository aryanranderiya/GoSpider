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
	maxURLs := flag.Int("urls", 1000, "Maximum number of URLs to process (default 1000). 0 = unlimited")
	numWorkers := flag.Int("workers", 5, "Number of concurrent workers (default 5)")
	useProxies := flag.Bool("proxies", false, "Use proxies from proxies.txt file")
	verbose := flag.Bool("verbose", false, "Enable verbose output (show found URLs and detailed processing info)")

	// Parse command line flags
	flag.Parse()

	// Validate required flags
	if *startURL == "" {
		fmt.Println("Error: -url flag is required")
		fmt.Println("Usage: go run main.go -url=https://example.com [-domains=3] [-urls=1000] [-workers=5] [-verbose]")
		flag.PrintDefaults()
		return
	}

	// Record start time
	startTime := time.Now()

	fmt.Println("\nwelcome to gospider by aryan randeriya")
	fmt.Printf("Starting URL: %s\n", *startURL)
	fmt.Printf("Max domains: %d\n", *maxDomains)
	fmt.Printf("Max URLs: %d\n", *maxURLs)
	fmt.Printf("Workers: %d\n", *numWorkers)
	fmt.Printf("Using proxies: %t\n", *useProxies)
	fmt.Printf("Verbose mode: %t\n", *verbose)

	// Load proxies if requested
	if *useProxies {
		if *verbose {
			fmt.Println("Loading proxies...")
		}
		utils.LoadProxies("proxies.txt", *verbose)
	}

	// Initialize your custom queue - this stores URLs waiting to be processed
	queue := internal.NewQueue(*maxDomains, *maxURLs, *verbose)
	queue.Enqueue(*startURL) // Add the starting URL to begin crawling

	// Create a channel to communicate URLs between main thread and worker threads
	urlChannel := make(chan string, 200) // buffered channel with larger capacity
	var wg sync.WaitGroup                // WaitGroup tracks how many workers are currently processing URLs

	// Start multiple worker goroutines (run in background)
	if *verbose {
		fmt.Printf("Starting %d workers...\n", *numWorkers)
	}
	for i := 0; i < *numWorkers; i++ {
		go internal.ProcessAllUrls(urlChannel, &wg, queue, *verbose)
	}

	// Progress reporting ticker (every 1 second)
	progressTicker := time.NewTicker(1 * time.Second)
	defer progressTicker.Stop()

	// Start progress reporting goroutine
	go func() {
		for range progressTicker.C {
			if !*verbose {
				elapsed := time.Since(startTime)
				processedCount := queue.ProcessedCount()
				domainsCount := queue.DomainsCount()

				// Calculate max values for display
				maxUrlsDisplay := "∞"
				if *maxURLs > 0 {
					maxUrlsDisplay = fmt.Sprintf("%d", *maxURLs)
				}

				fmt.Printf("Progress: %d/%s URLs processed, %d/%d domains, %d in queue (%.1fs)\n",
					processedCount, maxUrlsDisplay, domainsCount, *maxDomains, queue.Len(), elapsed.Seconds())
			}
		}
	}()

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
					if *verbose {
						if queue.IsCrawlingComplete() {
							fmt.Println("Crawling complete - domain limit reached and queue is empty.")
						} else {
							fmt.Println("All URLs processed. Shutting down...")
						}
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

	// Stop progress reporting
	progressTicker.Stop()

	// Show final progress update for non-verbose mode
	if !*verbose {
		elapsed := time.Since(startTime)
		processedCount := queue.ProcessedCount()
		domainsCount := queue.DomainsCount()

		// Calculate max values for display
		maxUrlsDisplay := "∞"
		if *maxURLs > 0 {
			maxUrlsDisplay = fmt.Sprintf("%d", *maxURLs)
		}

		fmt.Printf("Final: %d/%s URLs processed, %d/%d domains, %d in queue (%.1fs)\n",
			processedCount, maxUrlsDisplay, domainsCount, *maxDomains, queue.Len(), elapsed.Seconds())
	}

	// Calculate total execution time
	totalTime := time.Since(startTime)
	urlsPerSecond := float64(queue.ProcessedCount()) / totalTime.Seconds()

	// Print final statistics
	fmt.Printf("\n=== Crawling Complete ===\n")
	fmt.Printf("Total execution time: %.2fs\n", totalTime.Seconds())
	fmt.Printf("Total unique URLs discovered: %d\n", queue.VisitedCount())
	fmt.Printf("Total URLs processed: %d\n", queue.ProcessedCount())
	fmt.Printf("Total domains processed: %d\n", queue.DomainsCount())
	fmt.Printf("URLs remaining in queue: %d\n", queue.Len())
	fmt.Printf("Processing rate: %.2f URLs/second\n", urlsPerSecond)
}
