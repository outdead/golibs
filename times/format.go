package times

import (
	"fmt"
	"time"
)

// FmtDuration formats a time.Duration into a human-readable string with optional seconds.
// The duration is rounded to the nearest minute (or second) before formatting.
//
// The function provides two formatting modes:
// - Default: "[-]XXhXXm" format (hours and minutes)
// - With seconds: "[-]XXhXXmXXs" format (when secs[0] is true)
//
// Parameters:
//   - duration: The time duration to format (can be negative)
//   - secs: Optional boolean flag to include seconds in output
//
// Returns:
//   - Formatted duration string with leading zero padding and optional sign
func FmtDuration(duration time.Duration, withSeconds ...bool) string {
	sign := ""

	showSeconds := len(withSeconds) > 0 && withSeconds[0]

	if showSeconds {
		duration = duration.Round(time.Second)
	} else {
		duration = duration.Round(time.Minute)
	}

	if negative := duration < 0; negative {
		duration = -duration
		sign = "-"
	}

	hours := duration / time.Hour
	duration -= hours * time.Hour //nolint:durationcheck // Intentional duration math - safe conversion
	minutes := duration / time.Minute

	if !showSeconds {
		return fmt.Sprintf("%s%02dh%02dm", sign, hours, minutes)
	}

	duration -= minutes * time.Minute //nolint:durationcheck // Intentional duration math - safe conversion
	seconds := duration / time.Second

	return fmt.Sprintf("%s%02dh%02dm%02ds", sign, hours, minutes, seconds)
}
