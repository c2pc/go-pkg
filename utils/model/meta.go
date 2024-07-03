package model

type Meta[C any] struct {
	*Pagination[C]
	Filter
}

func NewMeta[C any](p *Pagination[C], f Filter) Meta[C] {
	return Meta[C]{p, f}
}
