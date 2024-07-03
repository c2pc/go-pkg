package model

type OrderBy = string

const (
	OrderByAsc  OrderBy = "ASC"
	OrderByDesc OrderBy = "DESC"
)

type Filter struct {
	OrderBy map[string]OrderBy
}

func NewFilter(orderBy map[string]OrderBy) Filter {
	return Filter{
		OrderBy: orderBy,
	}
}
