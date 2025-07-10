// Package jobticker provides a managed ticker for periodic job execution.
// It wraps time.Ticker with start/stop controls and error handling.
package jobticker

import (
	"sync"
	"sync/atomic"
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
	name    string
	quit    chan bool
	started atomic.Bool
	wg      sync.WaitGroup
	logger  Logger

	interval time.Duration
	handler  HandlerFunc

	stopTimeout time.Duration
	metrics     Metrics
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
		logger:   l,
		handler:  handler,
		interval: interval,
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
	if t.started.Load() {
		t.logger.Debug(t.name + ": already been started")

		return
	}

	t.quit = make(chan bool, 1)
	t.started.Store(true)

	t.wg.Add(1)

	go t.run()
}

// Stop initiates graceful shutdown of the ticker.
// Blocks until the current execution completes.
// Safe to call multiple times or on unstarted tickers.
func (t *Ticker) Stop() {
	if t.quit == nil || !t.started.Load() {
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

func (t *Ticker) IsRunning() bool {
	return t.started.Load()
}

// run contains the main ticker execution loop.
// Runs in a goroutine until stopped via quit channel.
func (t *Ticker) run() {
	defer func() {
		t.started.Store(false)
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
