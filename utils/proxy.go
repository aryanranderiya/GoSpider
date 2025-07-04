package utils

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	proxies     []string
	proxiesInit bool
	useProxies  bool
)

// ParseProxies reads and cleans the proxy file, returning only valid IP:PORT entries
func ParseProxies(filename string, verbose bool) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open proxy file: %v", err)
	}
	defer file.Close()

	var proxies []string
	var totalLines, validLines, duplicates int
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	// Regex to match IP:PORT format anywhere in the line
	ipPortRegex := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(\d+)`)

	if verbose {
		fmt.Printf("Parsing proxy file: %s\n", filename)
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		totalLines++

		// Skip empty lines
		if line == "" {
			continue
		}

		// Extract IP:PORT from anywhere in the line
		matches := ipPortRegex.FindStringSubmatch(line)
		if len(matches) >= 3 {
			ip := matches[1]
			port := matches[2]

			// Validate IP address
			if net.ParseIP(ip) != nil {
				proxy := fmt.Sprintf("%s:%s", ip, port)

				// Check for duplicates
				if seen[proxy] {
					duplicates++
					continue
				}

				seen[proxy] = true
				proxies = append(proxies, proxy)
				validLines++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Print statistics only in verbose mode
	if verbose {
		fmt.Printf("\n=== Proxy Parsing Results ===\n")
		fmt.Printf("Total lines processed: %d\n", totalLines)
		fmt.Printf("Valid proxies found: %d\n", validLines)
		fmt.Printf("Duplicates removed: %d\n", duplicates)
		fmt.Printf("Final proxy count: %d\n", len(proxies))
	}

	return proxies, nil
}

// WriteCleanProxies writes the cleaned proxy list to a new file
func WriteCleanProxies(proxies []string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	for _, proxy := range proxies {
		_, err := file.WriteString(proxy + "\n")
		if err != nil {
			return fmt.Errorf("failed to write proxy: %v", err)
		}
	}

	return nil
}

// GetRandomProxy returns a random proxy from the loaded list
func GetRandomProxy(verbose bool) string {
	if !proxiesInit {
		LoadProxies("proxies.txt", verbose)
	}

	if len(proxies) == 0 {
		return ""
	}

	return proxies[rand.Intn(len(proxies))]
}

// LoadProxies loads proxies from file into memory
func LoadProxies(filename string, verbose bool) error {
	if proxiesInit {
		return nil // Already loaded
	}

	useProxies = true // Enable proxy usage when loading
	parsed, err := ParseProxies(filename, verbose)
	if err != nil {
		if verbose {
			fmt.Printf("Warning: Could not load proxies: %v\n", err)
		}
		proxies = []string{} // Use empty list if loading fails
		useProxies = false   // Disable proxies if loading fails
	} else {
		proxies = parsed
		if verbose {
			fmt.Printf("\n=== Proxy Loading Complete ===\n")
			fmt.Printf("Successfully loaded %d proxies\n", len(proxies))
		}
	}

	proxiesInit = true
	return err
}

// CreateHTTPClientWithProxy creates an HTTP client with proxy fallback strategy
func CreateHTTPClientWithProxy(verbose bool) *http.Client {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Only use proxy if explicitly enabled and available
	if useProxies && len(proxies) > 0 {
		// Try up to 3 different proxies before falling back to direct connection
		maxRetries := 3
		for i := range maxRetries {
			proxy := GetRandomProxy(verbose)
			if proxy != "" {
				proxyURL, err := url.Parse("http://" + proxy)
				if err != nil {
					if verbose {
						fmt.Printf("Invalid proxy URL format, trying next: %s\n", proxy)
					}
					continue
				}

				// Test the proxy with a simple transport
				transport := &http.Transport{
					Proxy:               http.ProxyURL(proxyURL),
					IdleConnTimeout:     30 * time.Second,
					TLSHandshakeTimeout: 10 * time.Second,
				}

				client.Transport = transport
				if verbose {
					fmt.Printf("Using proxy (attempt %d/%d): %s\n", i+1, maxRetries, proxy)
				}
				return client
			}
		}
		if verbose {
			fmt.Printf("All proxy attempts failed, falling back to direct connection\n")
		}
	} else {
		if verbose {
			fmt.Println("Using direct connection (proxies disabled)")
		}
	}

	// Fallback to direct connection
	client.Transport = &http.Transport{
		IdleConnTimeout: 30 * time.Second,
	}
	return client
}

// testProxy tests if a proxy is working by making a simple HTTP request
func testProxy(proxyURL *url.URL) bool {
	transport := &http.Transport{
		Proxy:           http.ProxyURL(proxyURL),
		IdleConnTimeout: 5 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	// Test with a simple HTTP request to a reliable endpoint
	resp, err := client.Get("http://httpbin.org/ip")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// CreateHTTPClientWithTestedProxy creates an HTTP client with tested proxy fallback strategy
func CreateHTTPClientWithTestedProxy(verbose bool) *http.Client {
	client := &http.Client{
		Timeout: 30 * time.Second, // Increased timeout for slow servers
	}

	// Only use proxy if explicitly enabled and available
	if useProxies && len(proxies) > 0 {
		// Try up to 3 different proxies before falling back to direct connection
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			proxy := GetRandomProxy(verbose)
			if proxy != "" {
				proxyURL, err := url.Parse("http://" + proxy)
				if err != nil {
					if verbose {
						fmt.Printf("Invalid proxy URL format, trying next: %s\n", proxy)
					}
					continue
				}

				// Test the proxy before using it
				if verbose {
					fmt.Printf("Testing proxy (attempt %d/%d): %s\n", i+1, maxRetries, proxy)
				}
				if testProxy(proxyURL) {
					transport := &http.Transport{
						Proxy:                 http.ProxyURL(proxyURL),
						MaxIdleConns:          2000,
						MaxIdleConnsPerHost:   500,  // Increased for 1000 workers
						MaxConnsPerHost:       500,  // Increased for 1000 workers
						IdleConnTimeout:       90 * time.Second,
						TLSHandshakeTimeout:   10 * time.Second, // Increased for slow connections
						ResponseHeaderTimeout: 20 * time.Second, // Increased for slow responses
						ExpectContinueTimeout: 1 * time.Second,
						DisableKeepAlives:     false,
						DisableCompression:    false,
					}
					client.Transport = transport
					if verbose {
						fmt.Printf("✓ Proxy test successful, using: %s\n", proxy)
					}
					return client
				} else {
					if verbose {
						fmt.Printf("✗ Proxy test failed, trying next: %s\n", proxy)
					}
				}
			}
		}
		if verbose {
			fmt.Printf("All proxy tests failed, falling back to direct connection\n")
		}
	} else {
		if verbose {
			fmt.Println("Using direct connection (proxies disabled)")
		}
	}

	// Optimized direct connection with connection pooling
	client.Transport = &http.Transport{
		MaxIdleConns:          2000,
		MaxIdleConnsPerHost:   500,  // Increased for 1000 workers
		MaxConnsPerHost:       500,  // Increased for 1000 workers
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second, // Increased for slow connections
		ResponseHeaderTimeout: 20 * time.Second, // Increased for slow responses
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    false,
	}
	return client
}
