package clause

import (
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/i18n"
	"strings"
)

var (
	ErrOrderByUnknownColumn = apperr.New("order_by_unknown_column", apperr.WithTextTranslate(i18n.ErrOrderByUnknownColumn), apperr.WithCode(code.InvalidArgument))
)

type FieldOrderBy map[string]Order

type Order struct {
	Column string
	Join   string
}

func OrderByFilter(quoteTo func(string) string, orderBy map[string]string, fieldOrderBy FieldOrderBy) (string, []string, error) {
	var query []string
	var joins []string
	for k, v := range orderBy {
		if search, ok := fieldOrderBy[k]; ok {
			column := upperModels(quoteTo(search.Column))
			join := quoteTo(search.Join)
			order := OrderByAsc
			if strings.ToUpper(v) == OrderByDesc {
				order = OrderByDesc
			}
			query = append(query, fmt.Sprintf("%s %s", column, order))
			if join != "" {
				joins = append(joins, join)
			}
		} else {
			return "", nil, ErrOrderByUnknownColumn.WithTextArgs(k)
		}
	}

	return strings.Join(query, ","), joins, nil
}
