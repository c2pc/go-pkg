package request

import "github.com/gin-gonic/gin"

type PaginationRequest struct {
	Limit               int  `form:"limit" binding:"omitempty"`
	Offset              int  `form:"offset" binding:"omitempty"`
	MustReturnTotalRows bool `form:"count" binding:"omitempty"`
}

func Pagination(c *gin.Context) (*PaginationRequest, error) {
	return BindQuery[PaginationRequest](c)
}
