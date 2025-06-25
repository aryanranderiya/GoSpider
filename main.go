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

	// Initialize your custom queue
	queue := internal.NewQueue()
	queue.Enqueue(startURL)

	// Create a channel to hold URLs
	urlChan := make(chan string, 100) // buffered channel with capacity 100
	var wg sync.WaitGroup

	// Start one worker (you can add more)
	go func() {
		for url := range urlChan {
			fmt.Println("Processing:", url)
			time.Sleep(1 * time.Second) // simulate work
			wg.Done()
		}
	}()

	// Feed from the custom queue to the channel
	for {
		url, successfullyPopped := queue.Dequeue()
		if !successfullyPopped {
			break // queue is empty
		}

		wg.Add(1)
		urlChan <- url
	}

	wg.Wait()
	close(urlChan)
}
