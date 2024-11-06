package model

import "github.com/c2pc/go-pkg/v2/utils/clause"

type Filter struct {
	OrderBy map[string]string       `json:"orderBy"`
	Where   *clause.ExpressionWhere `json:"where"`
}

func NewFilter(orderBy map[string]string, where *clause.ExpressionWhere) Filter {
	return Filter{
		OrderBy: orderBy,
		Where:   where,
	}
}
