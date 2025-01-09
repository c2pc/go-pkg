package main

import (
	"context"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

type SSE struct {
	manager *SSEManager
}

func NewSSE(manager *SSEManager) *SSE {
	return &SSE{manager: manager}
}

func (s *SSE) InitHandler(api *gin.RouterGroup) {
	api.GET("/stream", sseHeadersMiddleware(), s.sseConnMiddleware(), func(c *gin.Context) {
		v, ok := c.Get("client")
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}
		client, ok := v.(Client)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		go func() {
			s.manager.Broadcast(Data{
				Message:  "new subscriber",
				From:     0,
				To:       0,
				Topic:    "info",
				PushType: PushTypeBackground,
			})
		}()

		c.Stream(func(w io.Writer) bool {
			select {
			case msg, ok := <-client.Channel:
				if !ok {
					return false
				}
				c.SSEvent("message", msg)
				return true

			case <-c.Request.Context().Done():
				return false
			}
		})

	})
}

func (s *SSE) Stream(c *gin.Context) {
	v, ok := c.Get("client")
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}
	client, ok := v.(Client)
	if !ok {
		c.Status(http.StatusInternalServerError)
		return
	}

	go func() {
		s.manager.Broadcast(Data{
			Message:  "New subscriber",
			From:     0,
			To:       0,
			Topic:    "info",
			PushType: PushTypeBackground,
		})
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-client.Channel:
			if !ok {
				return false
			}
			c.SSEvent("message", msg)
			return true

		case <-c.Request.Context().Done():
			return false
		}
	})
}

func (s *SSE) SendMessage(ctx context.Context, m Message) error {
	var fromID, toID int
	if m.From != nil {
		fromID = *m.From
	}
	if m.To != nil {
		toID = *m.To
	}

	data := Data{
		Message: m.Message,
		From:    fromID,
		To:      toID,
	}
	if m.Topic != nil {
		data.Topic = string(*m.Topic)
	}

	if m.PushType != nil {
		data.PushType = *m.PushType
	}

	select {
	case s.manager.broadcast <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func sseHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		c.Next()
	}
}

func (s *SSE) sseConnMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID, ok := mcontext.GetOpUserID(c.Request.Context())
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}

		client := Client{
			ID:      clientID,
			Channel: make(chan Data, s.manager.lenChan),
		}

		s.manager.RegisterClient(client)

		defer func() {
			s.manager.UnregisterClient(client)
		}()

		c.Set("client", client)
		c.Next()
	}
}
