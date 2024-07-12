package handler

import (
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
		zone.GET("", func(c *gin.Context) {})
		zone.POST("", func(c *gin.Context) {})
		zone.GET("/:book_id", func(c *gin.Context) {})
		zone.PATCH("/:book_id", func(c *gin.Context) {})
		zone.PUT("/:book_id", func(c *gin.Context) {})
		zone.DELETE("/:book_id", func(c *gin.Context) {})
	}
}
