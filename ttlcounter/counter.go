// Package ttlcounter provides a thread-safe TTL counter.
// All methods are safe for concurrent use by multiple goroutines.
// Could be useful for scenarios like rate limiting, tracking recent
// activity, or any case where you need to count events but only care
// about recent ones.
package ttlcounter

import (
	"container/heap"
	"sync"
	"time"
)

const (
	// DefaultTTL is the default time-to-live in seconds for counter items.
	DefaultTTL = 1 // Seconds

	// DefaultVacuumInterval is the default interval for automatic cleanup of expired items.
	DefaultVacuumInterval = 1 * time.Second
)

// Item represents a counter item with a value and last access timestamp.
type Item struct {
	value  int   // The current counter value
	access int64 // Unix timestamp of last access
}

// Counter is a TTL-based counter that automatically expires old entries.
type Counter struct {
	mu      sync.Mutex
	items   map[string]*Item
	ttl     int64 // Time-to-live in seconds for counter items
	stop    chan struct{}
	stopped sync.Once

	expQueue expirationQueue // Priority queue for expiration
}

// New creates a new Counter with the specified TTL (in seconds)
// and starts a background goroutine to periodically clean up expired items
// If ttl <= 0, DefaultTTL will be used.
func New(ttl int) *Counter {
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	counter := &Counter{
		items: make(map[string]*Item),
		ttl:   int64(ttl),
		stop:  make(chan struct{}),
	}

	heap.Init(&counter.expQueue)

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
	for k := range c.items {
		keys = append(keys, k)
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
	expireAt := now.Unix() + c.ttl

	item, ok := c.items[key]
	if !ok {
		item = &Item{}
		c.items[key] = item

		heap.Push(&c.expQueue, &expirationItem{
			key:      key,
			expireAt: expireAt,
		})
	}

	item.value++
	item.access = now.Unix()

	c.updateKeyInQueue(key, expireAt)
}

// Get returns the current value for the specified key
// Returns 0 if the key doesn't exist
// Note: This doesn't update the last access time.
func (c *Counter) Get(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	var value int
	if it, ok := c.items[key]; ok {
		value = it.value
	}

	return value
}

// Touch returns the current value for the specified key and resets last access time.
// Returns 0 if the key doesn't exist.
func (c *Counter) Touch(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expireAt := now.Unix() + c.ttl

	var value int
	if it, ok := c.items[key]; ok {
		value = it.value
		it.access = now.Unix()

		c.updateKeyInQueue(key, expireAt)
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

// Reset sets the counter value to 0 while keeping the key and updating access time.
func (c *Counter) Reset(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[key]; ok {
		item.value = 0
		item.access = time.Now().Unix()
	}
}

// Expire returns how many seconds until the key expires
// Returns a negative number if the key is already expired
// Returns 0 if the key doesn't exist.
func (c *Counter) Expire(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.items[key]; ok {
		return -int(time.Now().Unix() - v.access - c.ttl)
	}

	return 0
}

// Vacuum cleans up expired items based on the provided current time
// Called automatically by the background goroutine.
func (c *Counter) Vacuum(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for c.expQueue.Len() > 0 {
		// Peek at the next item to expire
		item := c.expQueue[0]
		if item.expireAt > now.Unix() {
			// No more items to expire
			break
		}

		// Remove from heap
		heap.Pop(&c.expQueue)

		// Only delete if not updated since expiration was scheduled
		if storedItem, exists := c.items[item.key]; exists && storedItem.access <= item.expireAt-c.ttl {
			delete(c.items, item.key)
		}
	}
}

// TTL returns the configured time-to-live in seconds.
func (c *Counter) TTL() int {
	return int(c.ttl)
}

// SetTTL updates the time-to-live for counter items
// If ttl <= 0, DefaultTTL will be used.
func (c *Counter) SetTTL(ttl int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl <= 0 {
		ttl = DefaultTTL
	}

	oldTTL := c.ttl
	c.ttl = int64(ttl)

	if oldTTL == c.ttl {
		return
	}

	newQueue := make(expirationQueue, 0, len(c.expQueue))

	// Update expiration time in queue
	for _, item := range c.expQueue {
		accessTime := item.expireAt - oldTTL
		newExpire := accessTime + c.ttl

		newQueue = append(newQueue, &expirationItem{
			key:      item.key,
			expireAt: newExpire,
		})
	}

	c.expQueue = newQueue
	heap.Init(&c.expQueue)
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
		case now := <-ticker.C:
			c.Vacuum(now)
		case <-c.stop:
			return
		}
	}
}

// updateKeyInQueue updates expiration time in queue for key.
func (c *Counter) updateKeyInQueue(key string, expireAt int64) {
	for i, eqItem := range c.expQueue {
		if eqItem.key == key {
			c.expQueue[i].expireAt = expireAt
			heap.Fix(&c.expQueue, i)

			break
		}
	}
}
