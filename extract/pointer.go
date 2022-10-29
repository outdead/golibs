package extract

// String extracts String value from pointer.
func String(p *string) string {
	if p != nil {
		return *p
	}

	return ""
}

// Int64 extracts Int64 value from pointer.
func Int64(p *int64) int64 {
	if p != nil {
		return *p
	}

	return 0
}
