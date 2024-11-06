package handler

import (
	"encoding/json"
	"net/http"

	"github.com/c2pc/go-pkg/v2/example/internal/model"
	"github.com/c2pc/go-pkg/v2/example/internal/service"
	"github.com/c2pc/go-pkg/v2/example/internal/transport/api/dto"
	"github.com/c2pc/go-pkg/v2/example/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/example/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/task"
	model3 "github.com/c2pc/go-pkg/v2/task/model"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type NewsHandler struct {
	newsService service.INewsService
	trx         mw.ITransaction
	taskService task.Tasker
}

func NewNewsHandlers(
	newsService service.INewsService,
	trx mw.ITransaction,
	taskService task.Tasker,
) *NewsHandler {
	return &NewsHandler{
		newsService,
		trx,
		taskService,
	}
}

func (h *NewsHandler) Init(api *gin.RouterGroup) {
	news := api.Group("/news")
	{
		news.POST(model3.MassDelete, h.taskService.MassDeleteHandler("news"))
		news.POST(model3.MassUpdate, h.taskService.MassUpdateHandler("news", h.MassUpdate))
		news.POST(model3.Import, h.taskService.ImportHandler("news", h.Import))
		news.POST(model3.Export, h.taskService.ExportHandler("news"))

		news.GET("", h.List)
		news.GET("/:id", h.GetById)
		news.POST("", h.trx.DBTransaction, h.Create)
		news.PATCH("/:id", h.trx.DBTransaction, h.Update)
		news.DELETE("/:id", h.trx.DBTransaction, h.Delete)
	}
}

func (h *NewsHandler) List(c *gin.Context) {
	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := model2.NewMeta(
		model2.NewPagination[model.News](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		model2.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.newsService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.NewsListTransform(c, m.Pagination))
}

func (h *NewsHandler) GetById(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.newsService.GetById(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.NewsTransform(data))
}

func (h *NewsHandler) Create(c *gin.Context) {
	cred, err := request2.BindJSON[request.NewsCreateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	news, err := h.newsService.Trx(request2.TxHandle(c)).Create(c.Request.Context(), dto.NewsCreate(cred))
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusCreated, transformer.NewsTransform(news))
}

func (h *NewsHandler) Update(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	cred, err := request2.BindJSON[request.NewsUpdateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	if err := h.newsService.Trx(request2.TxHandle(c)).Update(c.Request.Context(), id, dto.NewsUpdate(cred)); err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *NewsHandler) Delete(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	err = h.newsService.Trx(request2.TxHandle(c)).Delete(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *NewsHandler) MassUpdate(c *gin.Context) ([]byte, error) {
	cred, err := request2.BindJSON[request2.MultipleUpdateRequest[request.NewsMassUpdateRequest]](c)
	if err != nil {
		return nil, err
	}

	if cred == nil {
		return nil, apperr.ErrEmptyData
	}

	input := dto.NewsMassUpdate(*cred)

	return json.Marshal(input)
}

func (h *NewsHandler) Import(c *gin.Context) ([]byte, error) {
	cred, errs, err := request2.BindImportFileRequest[request.NewsImportRequest](c)
	if err != nil {
		return nil, err
	}

	input := dto.NewsImport(cred, errs)

	var ok bool
	if input.UserID, ok = mcontext.GetOpUserID(c.Request.Context()); !ok {
		return nil, apperr.ErrUnauthenticated
	}

	m, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	return m, nil
}
