package request

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/gin-gonic/gin"
)

type FilterRequest struct {
	OrderBy []clause.ExpressionOrderBy
	Where   *clause.ExpressionWhere
}

func Filter(c *gin.Context) (*FilterRequest, error) {
	type Filter struct {
		OrderBy string `form:"sorters" binding:"omitempty"`
		Where   string `form:"filters" binding:"omitempty"`
	}

	r, err := BindQuery[Filter](c)
	if err != nil {
		return nil, err
	}

	decodedWhere, err := url.QueryUnescape(r.Where)
	if err != nil {
		return nil, err
	}

	where, err := ParseWhere(decodedWhere)
	if err != nil {
		return nil, err
	}

	return &FilterRequest{
		OrderBy: orderByToMap(r.OrderBy),
		Where:   where,
	}, nil
}

func FilterJSON(c *gin.Context) (*FilterRequest, error) {
	type Filter struct {
		OrderBy string `json:"sorters" binding:"omitempty"`
		Where   string `json:"filters" binding:"omitempty"`
	}

	r, err := BindJSON[Filter](c)
	if err != nil {
		return nil, err
	}

	decodedWhere, err := url.QueryUnescape(r.Where)
	if err != nil {
		return nil, err
	}

	where, err := ParseWhere(decodedWhere)
	if err != nil {
		return nil, err
	}

	return &FilterRequest{
		OrderBy: orderByToMap(r.OrderBy),
		Where:   where,
	}, nil
}

func orderByToMap(str string) []clause.ExpressionOrderBy {
	var filter []clause.ExpressionOrderBy

	newStr := regexp.MustCompile(`[\[\] ]+`).ReplaceAllString(str, "")

	for _, v := range strings.Split(newStr, ",") {
		if len(v) > 0 {
			if strings.Contains(v, "-") {
				filter = append(filter, clause.ExpressionOrderBy{
					Column: v[1:],
					Order:  clause.OrderByDesc,
				})
			} else {
				filter = append(filter, clause.ExpressionOrderBy{
					Column: v,
					Order:  clause.OrderByDesc,
				})
			}
		}
	}

	return filter
}
