package model

type Pagination[C any] struct {
	Limit               int
	Offset              int
	MustReturnTotalRows bool
	TotalRows           int64
	Rows                []C
}

func NewPagination[C any](limit, offset int, mustReturnTotalRows bool) *Pagination[C] {
	return &Pagination[C]{Limit: limit, Offset: offset, MustReturnTotalRows: mustReturnTotalRows}
}

func (p *Pagination[C]) GetOffset() int {
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p.Offset
}

func (p *Pagination[C]) GetLimit() int {
	if p.Limit <= 0 {
		p.Limit = 0
	}
	return p.Limit
}
