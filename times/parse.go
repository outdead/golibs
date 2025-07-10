package times

import (
	"errors"
	"fmt"
	"time"
)

var ErrInvalidTimeFormat = errors.New("invalid time format")

// ParseOnlyTime parses a time string into hours, minutes, and seconds components.
// It supports multiple time formats with flexible parsing:
//   - "HH:MM:SS" (full time format)
//   - "HH:MM"    (hours and minutes, seconds default to 0)
//   - "HH"       (only hours, minutes and seconds default to 0)
//
// The function attempts each format in order and returns the first successful match.
// All returned time components are unsigned integers.
//
// Parameters:
//
//	strtime - string containing the time to parse (e.g., "14:30:00", "09:45", "23")
//
// Returns:
//
//	hour   - parsed hour (0-23)
//	minute - parsed minute (0-59)
//	second - parsed second (0-59)
//	err    - error if parsing fails, wrapping ErrInvalidTimeFormat with the invalid input
//
// Example usage:
//
//	h, m, s, err := ParseOnlyTime("08:30:15") // returns 8, 30, 15, nil
//	h, m, s, err := ParseOnlyTime("12:45")    // returns 12, 45, 0, nil
//	h, m, s, err := ParseOnlyTime("invalid")  // returns 0, 0, 0, error.
func ParseOnlyTime(strtime string) (hour, minute, second uint, err error) {
	formats := []string{
		"15:04:05", // Full format (HH:MM:SS)
		"15:04",    // Hours and minutes (HH:MM)
		"15",       // Only hours (HH)
	}

	for _, layout := range formats {
		if t, err := time.Parse(layout, strtime); err == nil {
			// #nosec G115 - Conversion is safe because time.Time methods
			// always return valid ranges (Hour: 0-23, Minute/Second: 0-59)
			return uint(t.Hour()), uint(t.Minute()), uint(t.Second()), nil
		}
	}

	return 0, 0, 0, fmt.Errorf("%w: %q", ErrInvalidTimeFormat, strtime)
}

// ParseOnlyTimeSafe tries to parse time string, returns zero values (0,0,0) on failure.
// Silent version that never fails - ideal for non-critical time parsing.
func ParseOnlyTimeSafe(strtime string) (hour, minute, second uint) {
	hour, minute, second, _ = ParseOnlyTime(strtime)

	return
}
