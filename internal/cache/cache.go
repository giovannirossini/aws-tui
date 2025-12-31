package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a cached value with TTL
type CacheEntry struct {
	Value      interface{}
	ExpiresAt  time.Time
	LastUpdate time.Time
}

// Cache is a thread-safe TTL-based cache
type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
}

// New creates a new cache instance
func New() *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
	}
}

// Set stores a value with the given TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	c.entries[key] = CacheEntry{
		Value:      value,
		ExpiresAt:  now.Add(ttl),
		LastUpdate: now,
	}
}

// Get retrieves a value if it exists and hasn't expired
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		// Entry expired, return miss (will be cleaned up by background job)
		return nil, false
	}

	return entry.Value, true
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// DeletePrefix removes all keys with the given prefix
func (c *Cache) DeletePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key := range c.entries {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			delete(c.entries, key)
		}
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]CacheEntry)
}

// CleanExpired removes expired entries (should be called periodically)
func (c *Cache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}

// Size returns the number of cached entries
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// GetAge returns the age of a cached entry
func (c *Cache) GetAge(key string) (time.Duration, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return 0, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return 0, false
	}

	return time.Since(entry.LastUpdate), true
}
