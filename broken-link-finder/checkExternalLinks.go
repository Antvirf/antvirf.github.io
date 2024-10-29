package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func checkExternalLink(url ExternalLink, results chan<- string, wg *sync.WaitGroup) {
	debugPrintf("Checking link: %s", url)
	defer wg.Done()
	resp, err := http.Get(url.externalUrl)
	if err != nil {
		results <- fmt.Sprintf("ERROR\t%s\t%s", url.externalUrl, err)
	} else {
		if !((resp.StatusCode == 200) && (hideSuccesses)) {
			results <- fmt.Sprintf("%d\t%s", resp.StatusCode, url.externalUrl)
		}
	}
}

func CheckExternalLinks(urls []ExternalLink) {
	var wg sync.WaitGroup
	results := make(chan string)
	counter := 0
	startTime := time.Now()
	for _, url := range urls {
		if wasInCache := cache.addToCache(url.externalUrl); !wasInCache {
			wg.Add(1)
			counter += 1
			go checkExternalLink(url, results, &wg)
		}
	}

	go func() {
		wg.Wait()      // Wait for all goroutines to finish
		close(results) // Close the result channel
	}()

	for result := range results {
		fmt.Println(result)
	}
	elapsed := time.Since(startTime)
	log.Printf("Processed %d links in %s (%f links per second)\n\n", counter, elapsed, float64(counter)/float64(elapsed.Seconds()))
}
