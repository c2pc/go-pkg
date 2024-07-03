package model

type Pagination[C any] struct {
	Limit     int
	Offset    int
	Count     bool
	TotalRows int64
	Rows      []C
}

func NewPagination[C any](limit, offset int, count bool) *Pagination[C] {
	return &Pagination[C]{Limit: limit, Offset: offset, Count: count}
}

func (p *Pagination[C]) GetOffset() int {
	if p.Offset < 0 {
		p.Offset = 0
	}
	return p.Offset
}

func (p *Pagination[C]) SetOffset(offset int) *Pagination[C] {
	p.Offset = offset
	return p
}

func (p *Pagination[C]) GetLimit() int {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	return p.Limit
}

func (p *Pagination[C]) SetLimit(limit int) *Pagination[C] {
	p.Limit = limit
	return p
}

func (p *Pagination[C]) GetCount() bool {
	return p.Count
}

func (p *Pagination[C]) setCount(count bool) *Pagination[C] {
	p.Count = count
	return p
}

func (p *Pagination[C]) GetRows() []C {
	return p.Rows
}

func (p *Pagination[C]) SetRows(rows []C) *Pagination[C] {
	p.Rows = rows
	return p
}

func (p *Pagination[C]) GetTotalRows() int64 {
	return p.TotalRows
}

func (p *Pagination[C]) SetTotalRows(totalRows int64) *Pagination[C] {
	p.TotalRows = totalRows
	return p
}
