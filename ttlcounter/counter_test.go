package ttlcounter

import (
	"sync"
	"testing"
	"time"
)

func TestCounterBasicOperations(t *testing.T) {
	c := New(2) // TTL 2 seconds

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
	c := New(3)

	c.Inc("test")
	initialExpire := c.Expire("test")

	time.Sleep(1 * time.Second)

	if val := c.Touch("test"); val != 1 {
		t.Fatalf("Touch returned wrong value: %d", val)
	}

	newExpire := c.Expire("test")
	if newExpire != initialExpire {
		t.Errorf("Expiration not updated by touch: was %d, now %d", initialExpire, newExpire)
	}

	if val := c.Get("test"); val != 1 {
		t.Errorf("Value changed after touch: %d", val)
	}
}

func TestCounterReset(t *testing.T) {
	c := New(10)

	c.Inc("test")
	c.Inc("test")
	c.Reset("test")

	if val := c.Get("test"); val != 0 {
		t.Errorf("Expected 0 after reset, got %d", val)
	}

	if c.Expire("test") <= 0 {
		t.Error("Reset should update access time")
	}
}

func TestCounterExpire(t *testing.T) {
	c := New(1)

	c.Inc("test")
	time.Sleep(1100 * time.Millisecond) // Slightly more than TTL

	if val := c.Get("test"); val != 0 {
		t.Errorf("Expected item to expire, got %d", val)
	}
}

func TestCounterConcurrentAccess(t *testing.T) {
	c := New(10)
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
	c := New(10)

	keys := []string{"a", "b", "c"}
	for _, k := range keys {
		c.Inc(k)
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

func TestCounterTTLUpdate(t *testing.T) {
	c := New(10)
	c.Inc("test")

	c.SetTTL(1)
	time.Sleep(1100 * time.Millisecond)

	if val := c.Get("test"); val != 0 {
		t.Errorf("Expected item to expire after TTL change, got %d", val)
	}
}

func TestCounterClose(t *testing.T) {
	c := New(1)
	c.Inc("test")

	c.Close()
	time.Sleep(1100 * time.Millisecond)

	// Vacuum should not run after close, so item should still exist
	if val := c.Get("test"); val != 1 {
		t.Errorf("Expected item to persist after close, got %d", val)
	}
}

func TestCounterDel(t *testing.T) {
	c := New(10)
	c.Inc("test")

	c.Del("test")
	if val := c.Get("test"); val != 0 {
		t.Errorf("Expected 0 after delete, got %d", val)
	}

	// Deleting non-existent key should not panic
	c.Del("nonexistent")
}
