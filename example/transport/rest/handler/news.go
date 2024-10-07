package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type NewsHandler struct {
}

func NewNewsHandlers() *NewsHandler {
	return &NewsHandler{}
}

func (h *NewsHandler) Init(api *gin.RouterGroup) {
	zone := api.Group("")
	{
		zone.GET("", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.POST("", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.GET("/:news_id", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.PATCH("/:news_id", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.PUT("/:news_id", func(c *gin.Context) { c.Status(http.StatusOK) })
		zone.DELETE("/:news_id", func(c *gin.Context) { c.Status(http.StatusOK) })
	}
}
