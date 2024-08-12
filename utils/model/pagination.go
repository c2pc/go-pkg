package model

const LIMIT = 10

type Pagination[C any] struct {
	Limit               int
	Offset              int
	MustReturnTotalRows bool
	TotalRows           int64
	Rows                []C
}

func NewPagination[C any](limit, offset int, mustReturnTotalRows bool) *Pagination[C] {
	if limit <= 0 {
		limit = LIMIT
	}

	if offset <= 0 {
		offset = 0
	}

	return &Pagination[C]{Limit: limit, Offset: offset, MustReturnTotalRows: mustReturnTotalRows}
}

func (p *Pagination[C]) GetOffset() int {
	return p.Offset
}

func (p *Pagination[C]) GetLimit() int {
	return p.Limit
}
