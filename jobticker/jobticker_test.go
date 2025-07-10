package jobticker

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

// MockLogger implements Logger interface for testing
type MockLogger struct {
	debugs []string
	errors []string
	mu     sync.Mutex
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg := ""
	if len(args) > 0 {
		msg = args[0].(string)
		if len(args) > 1 {
			msg += " " + fmt.Sprint(args[1])
		}
	}

	m.debugs = append(m.debugs, msg)
}

func (m *MockLogger) Error(args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	msg := ""
	if len(args) > 0 {
		msg = args[0].(string)
		if len(args) > 1 {
			msg += " " + fmt.Sprint(args[1])
		}
	}

	m.errors = append(m.errors, msg)
}

// MockMetrics implements Metrics interface for testing
type MockMetrics struct {
	observations []struct {
		Start  time.Time
		Delta  time.Duration
		Status bool
	}
}

func (m *MockMetrics) Observe(start time.Time, delta time.Duration, err error) {
	m.observations = append(m.observations, struct {
		Start  time.Time
		Delta  time.Duration
		Status bool
	}{start, delta, err == nil})
}

func TestTickerLifecycle(t *testing.T) {
	logger := &MockLogger{}

	var handlerCalls int

	handler := func() error {
		handlerCalls++

		return nil
	}

	ticker := New("test", handler, 10*time.Millisecond, logger)

	if ticker.IsRunning() {
		t.Error("Ticker should not be running after creation")
	}

	ticker.Start()
	time.Sleep(25 * time.Millisecond) // Allow 2-3 ticks

	if !ticker.IsRunning() {
		t.Error("Ticker should be running after Start()")
	}

	ticker.Stop()
	time.Sleep(15 * time.Millisecond) // Ensure stop completes

	if ticker.IsRunning() {
		t.Error("Ticker should not be running after Stop()")
	}

	if handlerCalls < 2 {
		t.Errorf("Expected at least 2 handler calls, got %d", handlerCalls)
	}
}

func TestTickerErrorHandling(t *testing.T) {
	logger := &MockLogger{}

	handler := func() error {
		return errors.New("test error")
	}

	ticker := Start("test", handler, 10*time.Millisecond, logger)
	time.Sleep(25 * time.Millisecond)
	ticker.Stop()

	if len(logger.errors) == 0 {
		t.Error("Expected error logs from handler")
	} else if logger.errors[0] != "test: test error" {
		t.Errorf("Unexpected error message: %v", logger.errors[0])
	}
}

func TestTickerPanicRecovery(t *testing.T) {
	logger := &MockLogger{}

	handler := func() error {
		panic("test panic")
	}

	ticker := New("test", handler, 10*time.Millisecond, logger)
	ticker.Start()
	time.Sleep(25 * time.Millisecond)
	ticker.Stop()

	if len(logger.errors) == 0 || logger.errors[0] != "test: handler panic: test panic" {
		t.Error("Expected panic recovery log")
	}
}

func TestTickerConcurrentControl(t *testing.T) {
	logger := &MockLogger{}
	handler := func() error { return nil }

	ticker := New("test", handler, 10*time.Millisecond, logger)

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			ticker.Start()
			ticker.Stop()
		}()
	}

	wg.Wait()

	// Should not panic and maintain correct state
	if ticker.IsRunning() {
		t.Error("Ticker should not be running after all stops")
	}
}

func TestTickerTimeout(t *testing.T) {
	logger := &MockLogger{}

	block := make(chan struct{})

	handler := func() error {
		<-block // Block forever

		return nil
	}

	ticker := New("test", handler, 10*time.Millisecond, logger, WithStopTimeout(5*time.Millisecond))
	ticker.Start()
	time.Sleep(11 * time.Millisecond) // Let it start

	start := time.Now()
	ticker.Stop()
	elapsed := time.Since(start)

	if elapsed < 5*time.Millisecond || elapsed > 6*time.Millisecond {
		t.Errorf("Unexpected stop duration: %v", elapsed)
	}

	if len(logger.errors) == 0 || logger.errors[0] != "test: forced shutdown due to timeout" {
		t.Error("Expected timeout error log")
	}

	ticker.Stop()

	close(block) // Cleanup
}

func TestTickerMetrics(t *testing.T) {
	metrics := &MockMetrics{}
	logger := &MockLogger{}

	handler := func() error { return nil }

	ticker := New("test", handler, 10*time.Millisecond, logger, WithMetrics(metrics))
	ticker.Start()
	time.Sleep(15 * time.Millisecond)
	ticker.Stop()

	if len(metrics.observations) == 0 {
		t.Error("Expected metrics observations")
	}
}

func TestTickerMultipleStart(t *testing.T) {
	logger := &MockLogger{}
	handler := func() error { return nil }

	ticker := New("test", handler, 10*time.Millisecond, logger)
	ticker.Start()
	ticker.Start() // Second call
	time.Sleep(5 * time.Millisecond)

	if len(logger.debugs) == 0 || logger.debugs[0] != "test: already been started" {
		t.Error("Expected duplicate start warning")
	}

	ticker.Stop()
}

func TestTickerMultipleStop(t *testing.T) {
	logger := &MockLogger{}
	handler := func() error { return nil }

	ticker := New("test", handler, 10*time.Millisecond, logger)
	ticker.Start()
	time.Sleep(5 * time.Millisecond)
	ticker.Stop()
	ticker.Stop() // Second call

	if len(logger.debugs) != 2 {
		t.Error("Unexpected debug logs", fmt.Sprintf(`["%s"]`, strings.Join(logger.debugs, `", "`)))
	}

	if logger.debugs[1] != "test: is not running" {
		t.Error("Expected duplicate stop warning")
	}
}
