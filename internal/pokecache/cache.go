// Package pokecache provides in-memory caching for HTTP responses.
// Placed under internal/ so only THIS module can import it.
package pokecache

import (
	"sync" // provides sync.Mutex for thread-safe map access
	"time" // provides time.Time, time.Duration, time.Ticker
)

// cacheEntry is one slot in the cache.
// Unexported (lowercase) — callers never touch this directly.
type cacheEntry struct {
	createdAt time.Time // timestamp of when Add() was called for this entry
	val       []byte    // the raw HTTP response body stored as bytes

}

// Cache is what callers create and use.
type Cache struct {
	entries  map[string]cacheEntry // key=URL, value=bytes+timestamp
	mu       sync.Mutex            // protects the map across goroutines
	interval time.Duration         // max age before an entry is deleted
}

// NewCache creates a Cache and immediately starts the background reaper.
// Call this once in main() and pass the pointer into your config.
func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		entries:  make(map[string]cacheEntry), // make() is required — a nil map would panic on write
		interval: interval,
	}
	go c.reapLoop() // 'go' spawns a goroutine — runs concurrently, doesn't block main
	return c
}

// Add stores a new entry. key = URL, val = raw HTTP response body.
func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()         // acquire the lock — no other goroutine can touch entries now
	defer c.mu.Unlock() // release the lock when this function returns (even if it panics)
	c.entries[key] = cacheEntry{
		createdAt: time.Now(), // record exactly when this was inserted
		val:       val,
	}
}

// Get returns the cached value for key.
// Returns (data, true) on hit, (nil, false) on miss.
// Always check the bool before using the data.
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false // cache miss
	}
	return entry.val, true // cache hit
}

// reapLoop runs forever in its goroutine, waking every interval to remove stale entries.
func (c *Cache) reapLoop() {
	ticker := time.NewTicker(c.interval) // fires every interval duration
	defer ticker.Stop()                  // clean up when/if this goroutine ever exits

	// ticker.C is a channel. for-range blocks here until the next tick arrives,
	// runs reap(), then goes back to waiting. Repeats forever.
	for range ticker.C {
		c.reap()
	}
}

// reap scans all entries and deletes any that are older than the interval.
func (c *Cache) reap() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		// now.Sub(entry.createdAt) = how long ago this entry was created
		if now.Sub(entry.createdAt) > c.interval {
			delete(c.entries, key) // safe to delete from a map while ranging over it in Go
		}
	}
}
