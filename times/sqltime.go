// Package times provides custom time handling utilities.
package times

import (
	"errors"
	"fmt"
	"time"
)

// DateTimeLayout defines the standard format for date-time strings
// Uses Go's reference time format: Mon Jan 2 15:04:05 MST 2006.
const DateTimeLayout = "2006-01-02 15:04:05"

// ErrUndefinedDateTime is returned when an unsupported type is encountered.
var ErrUndefinedDateTime = errors.New("undefined datetime")

// SQLTime is a custom type that wraps time.Time to provide custom scanning behavior.
type SQLTime time.Time

// Scan implements the sql.Scanner interface for SQLTime
// It converts various input types into a SQLTime value.
func (t *SQLTime) Scan(v interface{}) error {
	if v == nil {
		return nil
	}

	switch value := v.(type) {
	case time.Time:
		// Direct assignment if input is time.Time
		*t = SQLTime(value)
	case string:
		// Parse string using the defined layout
		vt, err := time.Parse(DateTimeLayout, value)
		if err != nil {
			return err
		}

		*t = SQLTime(vt)
	case []byte:
		// Handle byte slices (common in database drivers)
		vt, err := time.Parse(DateTimeLayout, string(value))
		if err != nil {
			return err
		}

		*t = SQLTime(vt)
	default:
		// Return error for unsupported types
		return fmt.Errorf("%w: %v", ErrUndefinedDateTime, value)
	}

	return nil
}
