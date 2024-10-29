package main

import (
	"net/http"
	"time"

	"golang.org/x/net/html"
)

func Crawl(url string, depth int, urls chan ExternalLink) {
	defer close(urls)
	if depth <= 0 {
		return
	}

	// Check if in cache, if yes then return immediately. If not, add immediately
	if wasInCache := cache.addToCache(url); wasInCache {
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		return
	}

	// Parse
	doc, err := html.Parse(resp.Body)
	defer resp.Body.Close()
	var links []string
	var externalLinks []string
	links, externalLinks = parseLinksFromResponse(doc, links, externalLinks)
	debugPrintf("[%s] external links:", url, externalLinks)

	// Send every current link back to main
	for i := range externalLinks {
		debugPrintf("[%s] sending to external: %s", url, externalLinks[i])
		urls <- ExternalLink{
			localPage:   url,
			externalUrl: externalLinks[i],
		}
	}

	// Go crawl more
	results := make([]chan ExternalLink, len(links))
	for i, u := range links {
		debugPrintf("start goroutine for link: %s", u)
		results[i] = make(chan ExternalLink)
		time.Sleep(time.Duration(scrapeDelay) * time.Microsecond)
		go Crawl(u, depth-1, results[i])
	}

	// Return stuff to main as it comes in
	for i := range results {
		for s := range results[i] {
			urls <- s
		}
	}
}
