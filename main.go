package main

import (
	"flag"
	"fmt"
	"gospider/internal"
	"gospider/utils"
	"io"
	"os"
	"sync"
	"time"
)

func main() {

	// Print ascii art
	f, _ := os.Open("ascii.txt")
	defer f.Close()
	data, _ := io.ReadAll(f)
	fmt.Println(string(data))

	// Define command line flags
	startURL := flag.String("url", "", "Starting URL to crawl (required)")
	maxDomains := flag.Int("domains", 100, "Maximum number of domains to crawl (default 100)")
	maxURLs := flag.Int("urls", 1000, "Maximum number of URLs to process (default 1000). 0 = unlimited")
	numWorkers := flag.Int("workers", 5, "Number of concurrent workers (default 5)")
	useProxies := flag.Bool("proxies", false, "Use proxies from proxies.txt file")
	downloadImages := flag.Bool("images", false, "Download images found during crawling")
	saveFiles := flag.Bool("save", false, "Save markdown files to disk (default false)")
	verbose := flag.Bool("verbose", false, "Enable verbose output (show found URLs and detailed processing info)")

	// Parse command line flags
	flag.Parse()

	// Validate required flags
	if *startURL == "" {
		fmt.Println("Error: -url flag is required")
		fmt.Println("Usage: go run main.go -url=https://example.com [-domains=3] [-urls=1000] [-workers=5] [-images] [-verbose]")
		flag.PrintDefaults()
		return
	}

	// Record start time
	startTime := time.Now()

	fmt.Println("\nwelcome to gospider by aryan randeriya \n\n")
	fmt.Printf("Starting URL: %s\n", *startURL)
	fmt.Printf("Max domains: %d\n", *maxDomains)
	fmt.Printf("Max URLs: %d\n", *maxURLs)
	fmt.Printf("Workers: %d\n", *numWorkers)
	fmt.Printf("Using proxies: %t\n", *useProxies)
	fmt.Printf("Download images: %t\n", *downloadImages)
	fmt.Printf("Save files: %t\n", *saveFiles)
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
	urlChannel := make(chan string, 10000) // much larger buffer for 1000 workers
	var wg sync.WaitGroup                  // WaitGroup tracks how many workers are currently processing URLs

	// Start multiple worker goroutines (run in background)
	if *verbose {
		fmt.Printf("Starting %d workers...\n", *numWorkers)
	}
	for i := 0; i < *numWorkers; i++ {
		go internal.ProcessAllUrls(urlChannel, &wg, queue, *downloadImages, *saveFiles, *verbose)
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
				completedCount := queue.CompletedCount()
				domainsCount := queue.DomainsCount()
				queueSize := queue.Len()

				// Calculate processing rate
				rate := float64(completedCount) / elapsed.Seconds()

				// Calculate max values for display
				maxUrlsDisplay := "∞"
				if *maxURLs > 0 {
					maxUrlsDisplay = fmt.Sprintf("%d", *maxURLs)
				}

				// Calculate percentage for domains
				domainPercent := float64(domainsCount) / float64(*maxDomains) * 100

				// Format time
				minutes := int(elapsed.Minutes())
				seconds := int(elapsed.Seconds()) % 60
				timeStr := fmt.Sprintf("%dm%ds", minutes, seconds)

				fmt.Printf("Processing: %d/%s URLs sent to workers | %d completed | %d/%d domains (%.1f%%) | %d queued | %.1f URLs/sec | %s\n",
					processedCount, maxUrlsDisplay, completedCount, domainsCount, *maxDomains, domainPercent, queueSize, rate, timeStr)
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
			time.Sleep(10 * time.Millisecond)

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
		completedCount := queue.CompletedCount()
		domainsCount := queue.DomainsCount()
		queueSize := queue.Len()

		// Calculate processing rate
		rate := float64(completedCount) / elapsed.Seconds()

		// Calculate max values for display
		maxUrlsDisplay := "∞"
		if *maxURLs > 0 {
			maxUrlsDisplay = fmt.Sprintf("%d", *maxURLs)
		}

		// Calculate percentage for domains
		domainPercent := float64(domainsCount) / float64(*maxDomains) * 100

		// Format time
		minutes := int(elapsed.Minutes())
		seconds := int(elapsed.Seconds()) % 60
		timeStr := fmt.Sprintf("%dm%ds", minutes, seconds)

		fmt.Printf("Final: %d/%s URLs sent to workers | %d completed | %d/%d domains (%.1f%%) | %d queued | %.1f URLs/sec | %s\n",
			processedCount, maxUrlsDisplay, completedCount, domainsCount, *maxDomains, domainPercent, queueSize, rate, timeStr)
	}

	// Calculate total execution time
	totalTime := time.Since(startTime)
	completedCount := queue.CompletedCount()
	processedCount := queue.ProcessedCount()
	visitedCount := queue.VisitedCount()
	domainsCount := queue.DomainsCount()
	queueSize := queue.Len()

	urlsPerSecond := float64(completedCount) / totalTime.Seconds()

	// Format execution time
	hours := int(totalTime.Hours())
	minutes := int(totalTime.Minutes()) % 60
	seconds := int(totalTime.Seconds()) % 60

	var timeDisplay string
	if hours > 0 {
		timeDisplay = fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		timeDisplay = fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		timeDisplay = fmt.Sprintf("%.1fs", totalTime.Seconds())
	}

	// Calculate success rate
	successRate := float64(completedCount) / float64(processedCount) * 100
	if processedCount == 0 {
		successRate = 0
	}

	// Print final statistics with better formatting
	fmt.Printf("\n=== Crawling Complete ===\n")
	fmt.Printf("Total execution time: %s\n", timeDisplay)
	fmt.Printf("Unique URLs discovered: %s\n", formatNumber(visitedCount))
	fmt.Printf("URLs sent to workers: %s\n", formatNumber(processedCount))
	fmt.Printf("URLs successfully processed: %s (%.1f%% success rate)\n", formatNumber(completedCount), successRate)
	fmt.Printf("Domains processed: %s\n", formatNumber(domainsCount))
	fmt.Printf("URLs remaining in queue: %s\n", formatNumber(queueSize))
	fmt.Printf("Processing rate: %.1f URLs/second\n", urlsPerSecond)
}

// formatNumber adds commas to large numbers for better readability
func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d", n/1000000, (n%1000000)/1000, n%1000)
}
