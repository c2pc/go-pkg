package request

import (
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/gin-gonic/gin"
	"regexp"
	"strings"
)

type FilterRequest struct {
	OrderBy map[string]string
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

	where, err := ParseWhere(r.Where)
	if err != nil {
		return nil, err
	}

	return &FilterRequest{
		OrderBy: orderByToMap(r.OrderBy),
		Where:   where,
	}, nil
}

func orderByToMap(str string) map[string]string {
	filter := map[string]string{}

	newStr := regexp.MustCompile(`[\[\] ]+`).ReplaceAllString(str, "")

	for _, v := range strings.Split(newStr, ",") {
		if len(v) > 0 {
			if strings.Contains(v, "-") {
				filter[v[1:]] = clause.OrderByDesc
			} else {
				filter[v] = clause.OrderByAsc
			}
		}
	}

	return filter
}
