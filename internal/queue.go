package internal

import (
	"fmt"
	"gospider/utils"
	"sync"
)

type Queue struct {
	urls       []string
	visited    map[string]bool
	domains    map[string]bool
	maxDomains int
	mu         sync.Mutex
}

func NewQueue(maxDomains int) *Queue {
	return &Queue{
		urls:       make([]string, 0),
		visited:    make(map[string]bool),
		domains:    make(map[string]bool),
		maxDomains: maxDomains,
	}
}

func (q *Queue) Enqueue(urlStr string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Skip if already visited
	if q.visited[urlStr] {
		return
	}

	// Extract domain from URL
	domain, valid := utils.ExtractDomain(urlStr)
	if !valid {
		return
	}

	// Check if we've reached max domains and this is a new domain
	if !q.domains[domain] && len(q.domains) >= q.maxDomains {
		fmt.Printf("Skipping new domain (max %d reached): %s\n", q.maxDomains, domain)
		return
	}

	// Track the domain (only if it's new and we haven't hit the limit)
	if !q.domains[domain] {
		q.domains[domain] = true
	}

	// Add to queue
	q.visited[urlStr] = true
	q.urls = append(q.urls, urlStr)
	fmt.Printf("Enqueued: %s (Queue size: %d, Domains: %d)\n", urlStr, len(q.urls), len(q.domains))
}

// Remove URL from the front
func (q *Queue) Dequeue() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.urls) == 0 {
		fmt.Println("Queue is empty. No more elements to dequeue (Queue size: 0)")
		return "", false
	}

	url := q.urls[0]
	q.urls = q.urls[1:]
	queueSize := len(q.urls)
	fmt.Printf("Dequeued: %s (Queue size: %d)\n", url, queueSize)
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

// Check if crawling should stop (domain limit reached and queue empty)
func (q *Queue) IsCrawlingComplete() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.domains) >= q.maxDomains && len(q.urls) == 0
}
