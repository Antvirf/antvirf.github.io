package main

import (
	"sync"
)

type Cache struct {
	values map[string]int
	mu     sync.Mutex
}

// Adds to cache if the value does not exist. Returns boolean denoting whether or not the entyr was ALREADY in the cache.
func (c *Cache) addToCache(value string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, ok := c.values[value]
	if ok {
		return true
	}
	c.values[value] = 1
	return false
}
