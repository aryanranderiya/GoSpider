package internal

import (
	"fmt"
	"gospider/utils"
	"sync"
)

type Queue struct {
	urls          []string
	visited       map[string]bool
	domains       map[string]bool
	maxDomains    int
	maxURLs       int
	processedURLs int
	mu            sync.Mutex
	verbose       bool
}

func NewQueue(maxDomains int, maxURLs int, verbose bool) *Queue {
	return &Queue{
		urls:          make([]string, 0),
		visited:       make(map[string]bool),
		domains:       make(map[string]bool),
		maxDomains:    maxDomains,
		maxURLs:       maxURLs,
		processedURLs: 0,
		verbose:       verbose,
	}
}

func (q *Queue) Enqueue(urlStr string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Skip if already visited
	if q.visited[urlStr] {
		return
	}

	// Check if we've reached max URLs limit
	if q.maxURLs > 0 && q.processedURLs >= q.maxURLs {
		if q.verbose {
			fmt.Printf("Skipping URL (max %d URLs reached): %s\n", q.maxURLs, urlStr)
		}
		return
	}

	// Extract domain from URL
	domain, valid := utils.ExtractDomain(urlStr, q.verbose)
	if !valid {
		return
	}

	// Check if we've reached max domains and this is a new domain
	if !q.domains[domain] && len(q.domains) >= q.maxDomains {
		if q.verbose {
			fmt.Printf("Skipping new domain (max %d reached): %s\n", q.maxDomains, domain)
		}
		return
	}

	// Track the domain (only if it's new and we haven't hit the limit)
	if !q.domains[domain] {
		q.domains[domain] = true
	}

	// Add to queue
	q.visited[urlStr] = true
	q.urls = append(q.urls, urlStr)
	if q.verbose {
		fmt.Printf("Enqueued: %s (Queue size: %d, Domains: %d, Processed: %d/%d)\n", urlStr, len(q.urls), len(q.domains), q.processedURLs, q.maxURLs)
	}
}

// Remove URL from the front
func (q *Queue) Dequeue() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if we've reached max URLs limit
	if q.maxURLs > 0 && q.processedURLs >= q.maxURLs {
		if q.verbose {
			fmt.Printf("Max URLs limit reached (%d/%d). Stopping crawl.\n", q.processedURLs, q.maxURLs)
		}
		return "", false
	}

	if len(q.urls) == 0 {
		if q.verbose {
			fmt.Println("Queue is empty. No more elements to dequeue (Queue size: 0)")
		}
		return "", false
	}

	url := q.urls[0]
	q.urls = q.urls[1:]
	q.processedURLs++
	queueSize := len(q.urls)
	if q.verbose {
		fmt.Printf("Dequeued: %s (Queue size: %d, Processed: %d/%d)\n", url, queueSize, q.processedURLs, q.maxURLs)
	}
	return url, true
}

// Length of the queue
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.urls)
}

// Check if a URL has been visited
func (q *Queue) HasVisited(urlStr string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.visited[urlStr]
}

// Get the total number of unique URLs visited
func (q *Queue) VisitedCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.visited)
}

// Get the total number of URLs processed
func (q *Queue) ProcessedCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.processedURLs
}

// Get the total number of domains discovered
func (q *Queue) DomainsCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.domains)
}

// Check if crawling should stop (domain limit reached and queue empty, or URL limit reached)
func (q *Queue) IsCrawlingComplete() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.maxURLs > 0 && q.processedURLs >= q.maxURLs {
		return true
	}
	return len(q.domains) >= q.maxDomains && len(q.urls) == 0
}
