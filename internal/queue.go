package internal

import (
	"container/list" // Doubly linked list package
)

// Queue of urls to fetch and scrape
type Queue struct {
	urls *list.List
}

// Create a new queue object
func NewQueue() *Queue {
	return &Queue{urls: list.New()}
}

// Insert url at the ending
func (q *Queue) Enqueue(url string) {
	q.urls.PushBack(url)
}

// Remove element from the front
func (q *Queue) Dequeue() (string, bool) {
	front := q.urls.Front()
	if front == nil {
		return "", false
	}
	q.urls.Remove(front)
	return front.Value.(string), true
}
