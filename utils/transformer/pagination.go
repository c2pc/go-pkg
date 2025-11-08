package transformer

import (
	"strconv"

	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/gin-gonic/gin"
)

const paginationHeader = "X-Total-Count"

func PaginationTransform[C any](c *gin.Context, p *meta.Pagination[C]) {
	if p.MustReturnTotalRows {
		c.Header(paginationHeader, strconv.FormatInt(p.TotalRows, 10))
	}
}
