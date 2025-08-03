package validator

import "errors"

var (
	// ErrNotFound is returned when data not found by identifier.
	ErrNotFound = errors.New("not found")

	// ErrDuplicateKey is returned when got duplicate key error from database.
	ErrDuplicateKey = errors.New("duplicate key")

	// ErrRequiredFieldMissed is returned when required param is empty in the request.
	ErrRequiredFieldMissed = errors.New("is required")
)

// Delimiters for text representation of errors.
const (
	ErrorSeparator = "; "
	FieldSeparator = " : "
)

// ValidationError contains field name and error message.
type ValidationError struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// NewValidationError creates and returns pointer to ValidationError.
func NewValidationError(field, msg string) *ValidationError {
	return &ValidationError{Name: field, Reason: msg}
}

// Error represents an error condition, with the nil value representing no error.
func (ve *ValidationError) Error() string {
	return ve.Name + FieldSeparator + ve.Reason
}

// ValidationErrors is an array of ValidationError's for use in custom error
// messages post validation.
type ValidationErrors []ValidationError

// Error represents an error condition, with the nil value representing no error.
func (ves ValidationErrors) Error() string {
	var message string

	for i, ve := range ves {
		message += ve.Error()

		if i+1 != len(ves) {
			message += ErrorSeparator
		}
	}

	return message
}
