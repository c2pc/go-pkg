package sse

import (
	"context"

	"github.com/c2pc/go-pkg/v2/sse/handler"
	"github.com/c2pc/go-pkg/v2/sse/model"
	"github.com/c2pc/go-pkg/v2/sse/service"
	"github.com/gin-gonic/gin"
)

type SSE interface {
	InitHandler(api *gin.RouterGroup)
	SendMessage(ctx context.Context, m model.Message) error
}

type sseImpl struct {
	sseHandler *handler.SSE
	sseManager *service.SSEManager
}

func New(lenChan int) SSE {
	sseManager := service.NewSSEManager(lenChan)

	sseHandler := handler.NewSSE(sseManager)

	return &sseImpl{
		sseHandler: sseHandler,
		sseManager: sseManager,
	}
}

func (s *sseImpl) InitHandler(api *gin.RouterGroup) {
	s.sseHandler.Init(api)
}

func (s *sseImpl) SendMessage(ctx context.Context, m model.Message) error {
	return s.sseManager.SendMessage(ctx, m)
}
