package extract

// String safely dereferences a string pointer, returning the string value or empty string if nil.
// This provides a nil-safe way to access string pointers and ensures you always get a valid string:
//   - If pointer is non-nil: returns the underlying string value
//   - If pointer is nil: returns an empty string ("")
//
// Useful when working with optional string fields in structs or API responses.
func String(pointer *string) string {
	if pointer != nil {
		return *pointer
	}

	return ""
}

// IsEmptyString checks if a string pointer is nil or points to an empty string.
// Returns true if either:
//   - The pointer is nil
//   - The dereferenced string is empty ("")
//
// Useful for safely checking optional string fields that may be nil.
func IsEmptyString(pointer *string) bool {
	return pointer == nil || *pointer == ""
}

// Int64 safely dereferences an int64 pointer, returning the value or 0 if nil.
//
// This helper function provides nil-safety when working with optional numeric fields:
//   - If pointer is non-nil: returns the underlying int64 value
//   - If pointer is nil: returns 0 (zero value for int64)
//
// Useful when working with optional int64 fields in structs or API responses.
func Int64(pointer *int64) int64 {
	if pointer != nil {
		return *pointer
	}

	return 0
}
