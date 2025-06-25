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
)

// ParseProxies reads and cleans the proxy file, returning only valid IP:PORT entries
func ParseProxies(filename string) ([]string, error) {
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

	fmt.Printf("Parsing proxy file: %s\n", filename)

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

	// Print statistics
	fmt.Printf("\n=== Proxy Parsing Results ===\n")
	fmt.Printf("Total lines processed: %d\n", totalLines)
	fmt.Printf("Valid proxies found: %d\n", validLines)
	fmt.Printf("Duplicates removed: %d\n", duplicates)
	fmt.Printf("Final proxy count: %d\n", len(proxies))

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
func GetRandomProxy() string {
	if !proxiesInit {
		LoadProxies("proxies.txt")
	}
	
	if len(proxies) == 0 {
		return ""
	}
	
	return proxies[rand.Intn(len(proxies))]
}

// LoadProxies loads proxies from file into memory
func LoadProxies(filename string) error {
	if proxiesInit {
		return nil // Already loaded
	}
	
	parsed, err := ParseProxies(filename)
	if err != nil {
		fmt.Printf("Warning: Could not load proxies: %v\n", err)
		proxies = []string{} // Use empty list if loading fails
	} else {
		proxies = parsed
		fmt.Printf("Loaded %d proxies\n", len(proxies))
	}
	
	proxiesInit = true
	rand.Seed(time.Now().UnixNano())
	return err
}

// CreateHTTPClientWithProxy creates an HTTP client with a random proxy
func CreateHTTPClientWithProxy() *http.Client {
	proxy := GetRandomProxy()
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	if proxy != "" {
		proxyURL, err := url.Parse("http://" + proxy)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
	}
	
	return client
}
