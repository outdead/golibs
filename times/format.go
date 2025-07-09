package times

import (
	"fmt"
	"time"
)

// FmtDuration formats a time.Duration into a standardized "HHhMMm" string representation.
// The duration is rounded to the nearest minute before formatting.
//
// Examples:
//
//	1h30m2s -> "01h30m"  (rounded from 1h30m2s)
//	25m     -> "00h25m"
//	3h45m   -> "03h45m"
//
// Parameters:
//
//	d: time.Duration to format (nanosecond precision)
//
// Returns:
//
//	Formatted string in "HHhMMm" format with leading zeros.
func FmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	h := d / time.Hour
	d -= h * time.Hour //nolint // TODO: Why no lint?
	m := d / time.Minute

	return fmt.Sprintf("%002dh%002dm", h, m)
}
