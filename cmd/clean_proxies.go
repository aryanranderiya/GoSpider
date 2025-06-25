package main

import (
	"flag"
	"fmt"
	"gospider/utils"
	"os"
)

func main() {
	inputFile := flag.String("input", "proxies.txt", "Input proxy file to clean")
	outputFile := flag.String("output", "clean_proxies.txt", "Output file for cleaned proxies")
	flag.Parse()

	fmt.Printf("Parsing proxy file: %s\n", *inputFile)

	proxies, err := utils.ParseProxies(*inputFile)
	if err != nil {
		fmt.Printf("Error parsing proxies: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d valid proxies\n", len(proxies))

	err = utils.WriteCleanProxies(proxies, *outputFile)
	if err != nil {
		fmt.Printf("Error writing clean proxies: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Clean proxies written to: %s\n", *outputFile)

	// Show first 10 proxies as preview
	fmt.Println("\nFirst 10 proxies:")
	for i, proxy := range proxies {
		if i >= 10 {
			break
		}
		fmt.Println(proxy)
	}
}
