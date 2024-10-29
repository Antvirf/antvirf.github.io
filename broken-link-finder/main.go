package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
)

var (
	scrapeDelay   int
	depth         int
	startingUrl   string
	hideSuccesses bool
	cache         Cache
	debugLogging  = false
)

type ExternalLink struct {
	localPage   string
	externalUrl string
	cacheKey    string
}

func main() {
	flag.StringVar(&startingUrl, "url", "http://localhost:1313", "Provide URL to scan.")
	flag.IntVar(&scrapeDelay, "delay", 500, "Delay in microseconds to start a new fetching goroutine.")
	flag.IntVar(&depth, "depth", 10, "Max recursion depth for mapping target site")
	flag.BoolVar(&hideSuccesses, "hide-successes", true, "Whether to hide successful link checks")
	flag.Parse()
	log.Printf("Starting broken links finder")
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("%s: %s\n", f.Name, f.Value)
	})

	// Prepare cache
	cache = Cache{
		mu:     sync.Mutex{},
		values: map[string]int{},
	}

	// Construct sitemap from local links only
	urlsToCheck := make(chan ExternalLink)
	go Crawl(startingUrl, depth, urlsToCheck)

	// Construct list of URLs
	urlsList := []ExternalLink{}
	for url := range urlsToCheck {
		urlsList = append(urlsList, url)
	}

	// Go check URLs
	CheckExternalLinks(urlsList)
}
