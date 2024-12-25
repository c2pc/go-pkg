package model

import "github.com/c2pc/go-pkg/v2/utils/clause"

type Filter struct {
	OrderBy []clause.ExpressionOrderBy `json:"orderBy"`
	Where   *clause.ExpressionWhere    `json:"where"`
}

func NewFilter(orderBy []clause.ExpressionOrderBy, where *clause.ExpressionWhere) Filter {
	return Filter{
		OrderBy: orderBy,
		Where:   where,
	}
}
