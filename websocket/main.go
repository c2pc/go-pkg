package websocket

import (
	"context"

	"github.com/gin-gonic/gin"
)

type WebSocket interface {
	InitHandler(api *gin.RouterGroup)
	SendMessage(ctx context.Context, m Message) error
	RegisterClientListener() <-chan Listener
	ShutDown()
}

type websocket struct {
	websocketHandler *handler
	websocketManager *manager
}

func New(lenChan int) WebSocket {
	websocketManager := newWebSocketManager(lenChan)
	websocketHandler := newWebSocket(websocketManager)

	return &websocket{
		websocketHandler: websocketHandler,
		websocketManager: websocketManager,
	}
}

func (s *websocket) InitHandler(api *gin.RouterGroup) {
	s.websocketHandler.Init(api)
}

func (s *websocket) SendMessage(ctx context.Context, m Message) error {
	return s.websocketManager.sendMessage(ctx, m)
}

func (s *websocket) ShutDown() {
	s.websocketManager.shutdown()
}

func (s *websocket) RegisterClientListener() <-chan Listener {
	return s.websocketManager.registerListener()
}
