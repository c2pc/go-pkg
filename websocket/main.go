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

func New(lenChan int, clientMaxCount ...int) WebSocket {
	var maxCount int
	if len(clientMaxCount) > 0 {
		maxCount = clientMaxCount[0]
	} else {
		maxCount = 10
	}

	websocketManager := newWebSocketManager(lenChan)
	websocketHandler := newWebSocket(websocketManager, maxCount)

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
