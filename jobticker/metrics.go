package jobticker

import "time"

type Metrics interface {
	Observe(start time.Time, duration time.Duration, err error)
}
