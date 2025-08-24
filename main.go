package main

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"sync"
)

type config struct {
	pages              map[string]int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxPages           int
}

func main() {
	args := os.Args

	if len(args) < 2 {
		fmt.Println("no website provided")
		os.Exit(1)
	}

	if len(args) > 5 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}
	rawBaseURL := args[1]

	maxConcurrency, err := strconv.Atoi(args[2])
	if err != nil {
		maxConcurrency = 10
		fmt.Println("invalid concurrency limit, using default value of 10")
	}
	maxPages, err := strconv.Atoi(args[3])
	if err != nil {
		maxPages = 100
		fmt.Println("invalid pages limit, using default value of 100")
	}

	fmt.Printf("starting crawl of: %v\n", rawBaseURL)

	baseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		fmt.Printf("Error parsing base URL: %v\n", err)
		os.Exit(1)
	}

	cfg := config{
		pages:              make(map[string]int),
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxConcurrency),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPages,
	}

	cfg.wg.Add(1)
	go cfg.crawlPage(rawBaseURL)
	cfg.wg.Wait()

	printReport(cfg.pages, rawBaseURL)
}

type pageCount struct {
	url   string
	count int
}

func printReport(pages map[string]int, baseURL string) {
	fmt.Println("=============================")
	fmt.Printf("  REPORT for %s\n", baseURL)
	fmt.Println("=============================")

	// Convert map to slice of structs for sorting
	pageSlice := make([]pageCount, 0, len(pages))
	for url, count := range pages {
		pageSlice = append(pageSlice, pageCount{url: url, count: count})
	}

	// Sort by count (descending), then by URL (ascending) for ties
	sort.Slice(pageSlice, func(i, j int) bool {
		if pageSlice[i].count == pageSlice[j].count {
			return pageSlice[i].url < pageSlice[j].url
		}
		return pageSlice[i].count > pageSlice[j].count
	})

	// Print each page
	for _, page := range pageSlice {
		fmt.Printf("Found %d internal links to %s\n", page.count, page.url)
	}
}
