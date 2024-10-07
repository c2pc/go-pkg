package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BookLikeHandler struct {
}

func NewBookLikeHandlers() *BookLikeHandler {
	return &BookLikeHandler{}
}

func (h *BookLikeHandler) Init(api *gin.RouterGroup) {
	zone := api.Group("")
	{
		zone.POST("/:book_id/likes", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.DELETE(":book_id/likes", func(c *gin.Context) { c.Status(http.StatusOK) })
	}
}
