package sse

import (
	"context"

	"github.com/c2pc/go-pkg/v2/sse/handlers"
	"github.com/c2pc/go-pkg/v2/sse/models"
	"github.com/c2pc/go-pkg/v2/sse/service"
	"github.com/gin-gonic/gin"
)

type SSE interface {
	InitHandler(api *gin.RouterGroup)
	SendMessage(ctx context.Context, messageType string, m models.Message) error
}

type sseImpl struct {
	sseHandler *handlers.SSE
}

func New(lenChan int) SSE {
	sseManager := service.NewSSEManager(lenChan)

	sseHandler := handlers.NewSSE(sseManager)

	return &sseImpl{
		sseHandler: sseHandler,
	}
}

func (s *sseImpl) InitHandler(api *gin.RouterGroup) {
	s.sseHandler.Init(api)
}

func (s *sseImpl) SendMessage(ctx context.Context, messageType string, m models.Message) error {
	return s.sseHandler.SendMessage(ctx, messageType, m)
}
