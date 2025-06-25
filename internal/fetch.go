package internal

import (
	"fmt"
	"gospider/utils"
	"io"
	"net/http"
	"sync"
)

func Fetch(url string, wg *sync.WaitGroup, queue *Queue) {
	wg.Add(1)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil) // method, url, body
	req.Header.Set("User-Agent", "GoSpider/1.0")

	response, error := client.Do(req)

	if error != nil {
		fmt.Println("Error while trying to fetch url: ", url, error)
		return
	}

	// always close the request body to prevent resource leakage
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading body:", err)
		return
	}

	fmt.Println("Successfully fetched url:", url)

	urls := utils.ExtractURLs(string(body))

	// Iterate over all urls and add to the queue for processing
	for _, url := range urls {
		fmt.Println("Found URL:", url)
		queue.Enqueue(url)
	}

	wg.Done()
}
