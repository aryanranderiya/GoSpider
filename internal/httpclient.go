package internal

import (
	"gospider/utils"
	"net/http"
	"sync"
)

var (
	httpClient     *http.Client
	httpClientOnce sync.Once
)

// GetHTTPClient returns a singleton HTTP client with optimized connection pooling
func GetHTTPClient(verbose bool) *http.Client {
	httpClientOnce.Do(func() {
		httpClient = utils.CreateHTTPClientWithTestedProxy(verbose)
	})
	return httpClient
}