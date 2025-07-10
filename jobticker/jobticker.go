// Package jobticker provides a managed ticker for periodic job execution.
// It wraps time.Ticker with start/stop controls and error handling.
package jobticker

import (
	"sync"
	"time"
)

// Logger describes the minimal logging interface required by the Ticker.
// Implementations should provide Debug and Error logging capabilities.
type Logger interface {
	Debug(args ...interface{})
	Error(args ...interface{})
}

// HandlerFunc defines the signature for functions that will be executed
// on each tick interval. Returning an error will log the failure.
type HandlerFunc func() error

// Ticker manages a recurring background job with start/stop capabilities.
// It ensures safe concurrent operation and proper resource cleanup.
type Ticker struct {
	name     string
	interval time.Duration
	handler  HandlerFunc
	logger   Logger
	metrics  Metrics

	wg          sync.WaitGroup
	stopTimeout time.Duration
	mu          sync.Mutex
	quit        chan bool
	started     bool
}

// New creates a configured but unstarted Ticker instance.
// Useful when you need to delay starting the ticker.
//
// Parameters:
//
//	name     - identifier for logging.
//	handler  - function to execute on each interval.
//	interval - time between executions.
//	l        - logger implementation.
//
// Returns:
//
//	*Ticker - started ticker instance (call Close when done).
func New(name string, handler HandlerFunc, interval time.Duration, l Logger, options ...Option) *Ticker {
	ticker := &Ticker{
		name:     name,
		interval: interval,
		handler:  handler,
		logger:   l,
	}

	for _, option := range options {
		option(ticker)
	}

	return ticker
}

// Start creates and immediately starts a new Ticker instance.
// This is the preferred entry point for most use cases.
//
// Parameters:
//
//	name     - identifier for logging.
//	handler  - function to execute on each interval.
//	interval - time between executions.
//	l        - logger implementation.
//
// Returns:
//
//	*Ticker - started ticker instance (call Close when done).
func Start(name string, handler HandlerFunc, interval time.Duration, l Logger, options ...Option) *Ticker {
	ti := New(name, handler, interval, l, options...)
	ti.Start()

	return ti
}

// Start begins the ticker's execution loop in a new goroutine.
// Safe to call multiple times (will log and ignore subsequent calls).
func (t *Ticker) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.started {
		t.logger.Debug(t.name + ": already been started")

		return
	}

	t.quit = make(chan bool, 1)
	t.started = true

	t.wg.Add(1)

	go t.run()
}

// Stop initiates a graceful shutdown of the ticker. It:
// 1. Sends a quit signal to the running goroutine
// 2. Waits for the current handler to complete
// 3. Implements timeout protection for stuck handlers
//
// Safe to call multiple times - will return immediately if:
// - Ticker isn't running (!started)
// - Quit channel is already full (shutdown in progress)
//
// Logs debug messages for all edge cases (already stopped, etc.).
func (t *Ticker) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.quit == nil || !t.started {
		t.logger.Debug(t.name + ": is not running")

		return
	}

	select {
	case t.quit <- true:
		if t.stopTimeout == 0 {
			t.wg.Wait() // waiting for goroutines

			return
		}

		done := make(chan struct{})
		go func() {
			t.wg.Wait() // waiting for goroutines
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(t.stopTimeout):
			t.logger.Error(t.name + ": forced shutdown due to timeout")
		}
	default:
		t.logger.Debug(t.name + ": close already been called")
	}
}

// IsRunning safely checks the ticker's current state.
// Returns:
//
//	true - if ticker is actively running
//	false - if stopped or never started
//
// Thread-safe atomic read - safe to call from any goroutine.
func (t *Ticker) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.started
}

// run is the main ticker event loop running in a goroutine.
// Handles:
// - Interval tick execution
// - Graceful shutdown signals
// - Resource cleanup on exit
//
// Note: Started by Start(), stopped by Stop().
// Uses defer for guaranteed state cleanup.
func (t *Ticker) run() {
	defer func() {
		t.started = false
		t.wg.Done()
	}()

	ticker := time.NewTicker(t.interval)

	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			t.executeHandler(now)
		case <-t.quit:
			t.logger.Debug(t.name + ": quit...")

			return
		}
	}
}

// executeHandler safely runs the user-provided handler function.
// Provides:
// - Panic recovery
// - Error logging
// - Performance metrics collection
//
// Parameters:
//
//	startedAt - timestamp when handler execution began.
func (t *Ticker) executeHandler(startedAt time.Time) {
	defer func() {
		if r := recover(); r != nil {
			t.logger.Error(t.name+": handler panic:", r)
		}
	}()

	err := t.handler()
	if err != nil {
		t.logger.Error(t.name+":", err)
	}

	if t.metrics != nil {
		t.metrics.Observe(startedAt, time.Since(startedAt), err)
	}
}
