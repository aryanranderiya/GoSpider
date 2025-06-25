package internal

import (
	"fmt"
	"sync"
)

type Queue struct {
	urls    []string
	visited map[string]bool // Track URLs we've already seen
	mu      sync.Mutex
}

// Create a new queue
func NewQueue() *Queue {
	return &Queue{
		urls:    make([]string, 0),
		visited: make(map[string]bool),
	}
}

// Add URL to the end of the queue
func (q *Queue) Enqueue(url string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check if URL has already been visited or queued
	if q.visited[url] {
		fmt.Printf("Skipping duplicate URL: %s (Queue size: %d)\n", url, len(q.urls))
		return
	}

	// Mark URL as visited and add to queue
	q.visited[url] = true
	q.urls = append(q.urls, url)
	queueSize := len(q.urls)
	fmt.Printf("Enqueued: %s (Queue size: %d)\n", url, queueSize)
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
func (q *Queue) HasVisited(url string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.visited[url]
}

// Get the total number of unique URLs visited
func (q *Queue) VisitedCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.visited)
}
