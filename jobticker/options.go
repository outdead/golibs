package jobticker

import "time"

type Option func(t *Ticker)

func WithStopTimeout(timeout time.Duration) Option {
	return func(t *Ticker) {
		t.stopTimeout = timeout
	}
}

func WithMetrics(metrics Metrics) Option {
	return func(t *Ticker) {
		t.metrics = metrics
	}
}
