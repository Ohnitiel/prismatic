package db

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type CacheEntry struct {
	Results   []map[string]any
	Timestamp time.Time
}

// Cache is a thread-safe in-memory cache
type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
	maxAge  time.Duration
}

func NewCache(maxAge time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
		maxAge:  maxAge,
	}
}

func (c *Cache) Set(connectionName string, query string, results []map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[getCacheKey(connectionName, query)] = CacheEntry{
		Results:   results,
		Timestamp: time.Now(),
	}
}

func (c *Cache) Get(connectionName string, query string) ([]map[string]any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[getCacheKey(connectionName, query)]
	if !ok {
		return nil, false
	}

	if time.Since(entry.Timestamp) > c.maxAge {
		slog.Info("Cache entry expired", "connection", connectionName, "query", query)
		return nil, false
	}

	return entry.Results, true
}

// Removes all cache entries
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]CacheEntry)
}

// Removes all cache entries older than the given duration
func (c *Cache) InvalidateOlder(olderThan time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	clearOlderThan := time.Now().Add(-olderThan)

	for key, entry := range c.entries {
		if entry.Timestamp.Before(clearOlderThan) {
			delete(c.entries, key)
		}
	}
}

// Returns the cache key hash (sha256) for the given connection and query
func getCacheKey(connectionName string, query string) string {
	data := fmt.Sprintf("%s-%s", connectionName, query)
	hash := sha256.Sum256([]byte(data))

	return fmt.Sprintf("%x", hash)
}
