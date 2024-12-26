package modelfilter

func ApplyLimits[T any](calls []T, offset int, limit int) []T {
	start := offset
	if start > len(calls) {
		start = len(calls)
	}
	end := start + limit
	if limit <= 0 || end > len(calls) {
		end = len(calls)
	}

	return calls[start:end]
}
