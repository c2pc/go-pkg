package transformer

func Nillable[C, T any](m *C, tr func(*C) *T) *T {
	if m == nil {
		return nil
	}

	return tr(m)
}

func NillableArray[C, T any](m []C, tr func(*C) *T) []*T {
	if len(m) == 0 {
		return nil
	}

	return Array(m, tr)
}

func Array[C, T any](m []C, tr func(*C) *T) []*T {
	var r []*T
	for _, a := range m {
		r = append(r, tr(&a))
	}

	return r
}
