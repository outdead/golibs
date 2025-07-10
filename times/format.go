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
//	1h30m2s -> "01h30m" (rounded from 1h30m2s)
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
func FmtDuration(duration time.Duration) string {
	sign := ""

	duration = duration.Round(time.Minute)

	if negative := duration < 0; negative {
		duration = -duration
		sign = "-"
	}

	h := duration / time.Hour
	duration -= h * time.Hour //nolint:durationcheck // Intentional duration math - safe conversion
	m := duration / time.Minute

	return fmt.Sprintf("%s%02dh%02dm", sign, h, m)
}
