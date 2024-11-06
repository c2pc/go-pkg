package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/task/internal/model"
	"github.com/c2pc/go-pkg/v2/task/internal/service"
	"github.com/c2pc/go-pkg/v2/task/internal/transport/api/transformer"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService service.ITaskService
	tr          mw.ITransaction
}

func NewTaskHandlers(
	taskService service.ITaskService,
	tr mw.ITransaction,
) *TaskHandler {
	return &TaskHandler{
		taskService,
		tr,
	}
}

func (h *TaskHandler) Init(api *gin.RouterGroup) {
	task := api.Group("tasks")
	{
		task.GET("", h.List)
		task.POST("/:id/stop", h.Stop)
		task.POST("/:id/rerun", h.Rerun)
		task.GET("/:id", h.GetById)
		task.GET("/:id/download", h.Download)
	}
}

func (h *TaskHandler) List(c *gin.Context) {
	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := model2.NewMeta(
		model2.NewPagination[model.Task](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		model2.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.taskService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.TaskListTransform(c, m.Pagination))
}

func (h *TaskHandler) GetById(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.taskService.GetById(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.TaskTransform(data))
}

func (h *TaskHandler) Stop(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	err = h.taskService.Stop(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *TaskHandler) Rerun(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.taskService.Rerun(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.SimpleTaskTransform(data))
}

func (h *TaskHandler) Download(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	path, err := h.taskService.Download(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.File(path)
}
