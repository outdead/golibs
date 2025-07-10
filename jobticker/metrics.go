package jobticker

import "time"

type Metrics interface {
	Observe(name string, start time.Time, duration time.Duration, err error)
}
