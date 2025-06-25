package internal

import (
	"fmt"
	"sync"
)

type Queue struct {
	urls []string
	mu   sync.Mutex
}

// Create a new queue
func NewQueue() *Queue {
	return &Queue{urls: make([]string, 0)}
}

// Add URL to the end of the queue
func (q *Queue) Enqueue(url string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.urls = append(q.urls, url)
	fmt.Printf("Enqueued: %s\n", url)
}

// Remove URL from the front
func (q *Queue) Dequeue() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.urls) == 0 {
		fmt.Println("Queue is empty. No more elements to dequeue")
		return "", false
	}

	url := q.urls[0]
	q.urls = q.urls[1:]
	fmt.Printf("Dequeued: %s\n", url)
	return url, true
}

// Length of the queue
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.urls)
}
