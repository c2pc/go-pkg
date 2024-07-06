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
	r := make([]*T, len(m))
	for i, a := range m {
		r[i] = tr(&a)
	}

	return r
}
