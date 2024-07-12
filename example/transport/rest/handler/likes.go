package handler

import (
	"github.com/gin-gonic/gin"
)

type LikeHandler struct {
}

func NewLikeHandlers() *LikeHandler {
	return &LikeHandler{}
}

func (h *LikeHandler) Init(api *gin.RouterGroup) {
	zone := api.Group("")
	{
		zone.GET("/:book_id/likes", func(c *gin.Context) {})
		zone.POST("/:book_id/likes", func(c *gin.Context) {})
		zone.GET("/:book_id/likes/:book_id", func(c *gin.Context) {})
		zone.PATCH("/:book_id/likes/:book_id", func(c *gin.Context) {})
		zone.PUT("/likes/:book_id", func(c *gin.Context) {})
		zone.DELETE("/likes/:book_id", func(c *gin.Context) {})
	}
}
