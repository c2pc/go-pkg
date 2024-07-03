package request

import (
	"github.com/gin-gonic/gin"
)

type MetaRequest struct {
	*PaginationRequest
	*FilterRequest
}

func Meta(c *gin.Context) (*MetaRequest, error) {
	var r MetaRequest
	var err error

	r.FilterRequest, err = Filter(c)
	if err != nil {
		return nil, err
	}

	r.PaginationRequest, err = Pagination(c)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
