package ttlcounter

import (
	"sync"
	"testing"
	"time"
)

func TestCounterBasicOperations(t *testing.T) {
	c := New(2 * time.Second) // TTL 2 seconds

	// Test Inc and Get
	c.Inc("test")
	if val := c.Get("test"); val != 1 {
		t.Errorf("Expected 1, got %d", val)
	}

	// Test multiple increments
	c.Inc("test")
	c.Inc("test")
	if val := c.Get("test"); val != 3 {
		t.Errorf("Expected 3, got %d", val)
	}

	// Test non-existent key
	if val := c.Get("nonexistent"); val != 0 {
		t.Errorf("Expected 0, got %d", val)
	}
}

func TestCounterTouch(t *testing.T) {
	c := New(3 * time.Second)

	c.Inc("test")
	initialExpire := c.Expire("test")

	time.Sleep(1 * time.Second)

	if val := c.Touch("test"); val != 1 {
		t.Fatalf("Touch returned wrong value: %d", val)
	}

	newExpire := c.Expire("test")
	if int(newExpire.Seconds()) != int(initialExpire.Seconds()) {
		t.Errorf("Expiration not updated by touch: was %v, now %v", initialExpire.Seconds(), newExpire.Seconds())
	}

	if val := c.Get("test"); val != 1 {
		t.Errorf("Value changed after touch: %d", val)
	}
}

func TestCounterExpire(t *testing.T) {
	c := New(1 * time.Second)

	c.Inc("test")

	time.Sleep(900 * time.Millisecond) // Slightly more than TTL

	if val := c.Get("test"); val != 1 {
		t.Errorf("Expected item, got %d", val)
	}

	time.Sleep(200 * time.Millisecond) // Slightly more than TTL

	if val := c.Get("test"); val != 0 {
		t.Errorf("Expected item to expire, got %d", val)
	}
}

func TestCounterConcurrentAccess(t *testing.T) {
	c := New(10 * time.Second)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Inc("concurrent")
			c.Get("concurrent")
			c.Touch("concurrent")
		}()
	}

	wg.Wait()

	if val := c.Get("concurrent"); val != 100 {
		t.Errorf("Expected 100 concurrent increments, got %d", val)
	}
}

func TestCounterKeys(t *testing.T) {
	c := New(10 * time.Second)

	keys := []string{"a", "b", "c"}
	for _, k := range keys {
		c.Inc(k)
	}

	if c.Len() != len(keys) {
		t.Errorf("Expected %d items, got %d", len(keys), c.Len())
	}

	retrieved := c.Keys()
	if len(retrieved) != len(keys) {
		t.Errorf("Expected %d keys, got %d", len(keys), len(retrieved))
	}

	// Convert to map for easier comparison
	keysMap := make(map[string]bool)
	for _, k := range retrieved {
		keysMap[k] = true
	}

	for _, k := range keys {
		if !keysMap[k] {
			t.Errorf("Key %s not found in returned keys", k)
		}
	}
}

func TestCounterTTL(t *testing.T) {
	c := New(0) // Sets to 1 by default
	c.Inc("test")

	if c.TTL() != DefaultTTL {
		t.Errorf("Expected default TTL, got %d", c.TTL())
	}

	// Vacuum should not run after close, so item should still exist
	if val := c.Get("test"); val != 1 {
		t.Errorf("Expected item to persist after close, got %d", val)
	}
}

func TestCounterTTLUpdate(t *testing.T) {
	c := New(10 * time.Second)

	if c.TTL() != 10*time.Second {
		t.Errorf("Expected TTL to be %d, got %d", 10, c.TTL())
	}

	c.Inc("test")

	c.SetTTL(-1)

	if c.TTL() != DefaultTTL {
		t.Errorf("Expected TTL to be %d, got %d", DefaultTTL, c.TTL())
	}
}

func TestCounterClose(t *testing.T) {
	c := New(1 * time.Second)
	c.Inc("test")

	c.Close()
	c.Close()
	time.Sleep(1100 * time.Millisecond)

	if val := c.Get("test"); val != 0 {
		t.Errorf("Expected empty item after close, got %d", val)
	}
}

func TestCounterDel(t *testing.T) {
	c := New(10 * time.Second)
	c.Inc("test")

	c.Del("test")
	if val := c.Get("test"); val != 0 {
		t.Errorf("Expected 0 after delete, got %d", val)
	}

	// Deleting non-existent key should not panic
	c.Del("nonexistent")
}
