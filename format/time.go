package format

import (
	"fmt"
	"time"
)

// FmtDuration converts time.Durations to format string.
func FmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour //nolint
	m := d / time.Minute

	return fmt.Sprintf("%002dh%002dm", h, m)
}
