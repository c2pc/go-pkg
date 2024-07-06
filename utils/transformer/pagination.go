package transformer

import (
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/gin-gonic/gin"
	"strconv"
)

const paginationHeader = "X-Total-Count"

func PaginationTransform[C any](c *gin.Context, p *model.Pagination[C]) {
	fmt.Println(paginationHeader, p.TotalRows)
	if p.MustReturnTotalRows {
		c.Header(paginationHeader, strconv.FormatInt(p.TotalRows, 10))
	}
}
