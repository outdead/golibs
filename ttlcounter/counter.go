// Package ttlcounter provides a thread-safe TTL counter.
// All methods are safe for concurrent use by multiple goroutines.
// Could be useful for scenarios like rate limiting, tracking recent
// activity, or any case where you need to count events but only care
// about recent ones.
package ttlcounter

import (
	"sync"
	"time"
)

const (
	// DefaultTTL is the default time-to-live in seconds for counter items.
	DefaultTTL = 1 * time.Second

	// DefaultVacuumInterval is the default interval for automatic cleanup of expired items.
	DefaultVacuumInterval = 1 * time.Second
)

// Item represents a counter item with a value and last access timestamp.
type Item struct {
	value  int   // The current counter value
	access int64 // Unix timestamp of last access
}

func (item *Item) Expired(ttl time.Duration) bool {
	return item.access+ttl.Nanoseconds() <= time.Now().UnixNano()
}

// Counter is a TTL-based counter that automatically expires old entries.
type Counter struct {
	mu      sync.Mutex
	items   map[string]*Item
	ttl     time.Duration
	stop    chan struct{}
	stopped sync.Once
}

// New creates a new Counter with the specified TTL (in seconds)
// and starts a background goroutine to periodically clean up expired items
// If ttl <= 0, DefaultTTL will be used.
func New(ttl time.Duration) *Counter {
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	counter := &Counter{
		items: make(map[string]*Item),
		ttl:   ttl,
		stop:  make(chan struct{}),
	}

	// Start a background goroutine to clean up expired items every second
	go counter.vacuumLoop()

	return counter
}

// Len returns the current number of items in the counter.
func (c *Counter) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.items)
}

// Keys returns a slice of all existing keys in the counter
// The order of keys is not guaranteed.
func (c *Counter) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}

	return keys
}

// Inc increments the counter for the specified key
// If the key doesn't exist, it creates a new counter starting at 1
// Updates the last access time to current time.
func (c *Counter) Inc(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	item, ok := c.items[key]
	if !ok {
		item = &Item{}
		c.items[key] = item
	}

	item.value++
	item.access = now.UnixNano()
}

// Get returns the current value for the specified key
// Returns 0 if the key doesn't exist
// Note: This doesn't update the last access time.
func (c *Counter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var value int

	if item, ok := c.items[key]; ok {
		if item.Expired(c.ttl) {
			return 0
		}

		value = item.value
	}

	return value
}

// Touch returns the current value for the specified key and resets last access time.
// Returns 0 if the key doesn't exist.
func (c *Counter) Touch(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()

	var value int

	if item, ok := c.items[key]; ok {
		if item.Expired(c.ttl) {
			return 0
		}

		value = item.value
		item.access = now.UnixNano()
	}

	return value
}

// Del removes the specified key from the counter.
// If the key doesn't exist, nothing happens.
func (c *Counter) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Expire returns how many seconds until the key expires
// Returns a negative number if the key is already expired
// Returns 0 if the key doesn't exist.
func (c *Counter) Expire(key string) time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()

	if item, ok := c.items[key]; ok {
		remaining := item.access + c.ttl.Nanoseconds() - now
		if remaining > 0 {
			return time.Duration(remaining)
		}

		return 0
	}

	return 0
}

// Vacuum cleans up expired items based on the provided current time
// Called automatically by the background goroutine.
func (c *Counter) Vacuum() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if item.Expired(c.ttl) {
			delete(c.items, key)
		}
	}
}

// TTL returns the configured time-to-live in seconds.
func (c *Counter) TTL() time.Duration {
	return c.ttl
}

// SetTTL updates the time-to-live for counter items
// If ttl <= 0, DefaultTTL will be used.
func (c *Counter) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl <= 0 {
		ttl = DefaultTTL
	}

	c.ttl = ttl
}

// Close stops the background cleanup goroutine
// Safe to call multiple times.
func (c *Counter) Close() {
	c.stopped.Do(func() {
		close(c.stop)
	})
}

// vacuumLoop runs in a goroutine and periodically calls Vacuum
// until the counter is closed.
func (c *Counter) vacuumLoop() {
	ticker := time.NewTicker(DefaultVacuumInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.Vacuum()
		case <-c.stop:
			return
		}
	}
}
