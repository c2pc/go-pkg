package transformer

import (
	"github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/gin-gonic/gin"
	"strconv"
)

const paginationHeader = "X-Total-Count"

func PaginationTransform[C any](c *gin.Context, p *model.Pagination[C]) {
	if p.Count {
		c.Request.Header.Set(paginationHeader, strconv.FormatInt(p.TotalRows, 10))
	}
}
