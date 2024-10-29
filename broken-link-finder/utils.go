package main

import (
	"errors"
	"log"
	"strings"

	"golang.org/x/net/html"
)

func debugPrintf(format string, v ...any) {
	if debugLogging {
		log.Printf(format, v)
	}
}

func getHrefFromAttrs(attributes []html.Attribute) (string, error) {
	for i := range attributes {
		if attributes[i].Key == "href" {
			return attributes[i].Val, nil
		}
	}
	return "", errors.New("no href found")
}

func deduplicate[T comparable](slicelist []T) []T {
	allKeys := make(map[T]bool)
	list := []T{}
	for _, item := range slicelist {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func parseLinksFromResponse(node *html.Node, links []string, externalLinks []string) ([]string, []string) {
	if node.Type == html.ElementNode && node.Data == "a" {
		link, err := getHrefFromAttrs(node.Attr)
		if err == nil {
			switch {
			case strings.HasPrefix(link, "/"):
				links = append(links, startingUrl+link)
			case !strings.HasPrefix(link, "#"):
				externalLinks = append(externalLinks, link)
			}
		}

	}
	// We only recurse down local pages
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		links, externalLinks = parseLinksFromResponse(c, links, externalLinks)
	}
	links = deduplicate(links)
	externalLinks = deduplicate(externalLinks)

	return links, externalLinks
}
