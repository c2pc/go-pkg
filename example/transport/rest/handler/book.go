package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BookHandler struct {
}

func NewBookHandlers() *BookHandler {
	return &BookHandler{}
}

func (h *BookHandler) Init(api *gin.RouterGroup) {
	zone := api.Group("")
	{
		zone.GET("", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.POST("", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.GET("/:book_id", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.PATCH("/:book_id", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.PUT("/:book_id", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.DELETE("/:book_id", func(c *gin.Context) { c.Status(http.StatusOK) })
	}
}
