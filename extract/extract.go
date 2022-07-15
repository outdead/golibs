package extract

func String(p *string) string {
	if p != nil {
		return *p
	}

	return ""
}

func Int64(p *int64) int64 {
	if p != nil {
		return *p
	}

	return 0
}
