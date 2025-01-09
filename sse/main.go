package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()

	manager := NewSSEManager(10)

	sse := NewSSE(manager)

	api := r.Group("/sse")
	sse.InitHandler(api)
	r.POST("/send", func(c *gin.Context) {
		var msg Message
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		if err := sse.SendMessage(c.Request.Context(), msg); err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	})

	r.StaticFile("/", "./index.html")

	r.Run(":8080")
}
