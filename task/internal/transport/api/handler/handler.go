package handler

import (
	"github.com/c2pc/go-pkg/v2/task/internal/service"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	"github.com/gin-gonic/gin"
)

type IHandler interface {
	Init(api *gin.RouterGroup)
}

type Handler struct {
	taskService service.ITaskService
	tr          mw.ITransaction
}

func NewHandlers(
	taskService service.ITaskService,
	tr mw.ITransaction,
) *Handler {
	return &Handler{
		taskService,
		tr,
	}
}

func (h *Handler) Init(api *gin.RouterGroup) {
	NewTaskHandlers(h.taskService, h.tr).Init(api)
}