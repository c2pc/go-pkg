package handler

import (
	"encoding/json"
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth"
	"github.com/c2pc/go-pkg/v2/example/internal/model"
	"github.com/c2pc/go-pkg/v2/example/internal/service"
	"github.com/c2pc/go-pkg/v2/example/internal/transport/api/dto"
	"github.com/c2pc/go-pkg/v2/example/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/example/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/task/types"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
)

type NewsHandler struct {
	news        service.INews
	trx         mw.ITransaction
	taskService auth.Tasker
}

func NewNewsHandlers(
	news service.INews,
	trx mw.ITransaction,
	taskService auth.Tasker,
) *NewsHandler {
	return &NewsHandler{
		news,
		trx,
		taskService,
	}
}

func (h *NewsHandler) Init(api *gin.RouterGroup) {
	news := api.Group("/news")
	{
		news.POST(types.Export, h.taskService.ExportHandler("news", h.Export))
		news.POST(types.Import, h.taskService.ImportHandler("news", h.Import))
		news.POST(types.MassUpdate, h.taskService.MassUpdateHandler("news", h.MassUpdate))
		news.POST(types.MassDelete, h.taskService.MassDeleteHandler("news", h.MassDelete))

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

	m := meta.NewMeta(
		meta.NewPagination[model.News](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		meta.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.news.List(c.Request.Context(), &m); err != nil {
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

	data, err := h.news.GetById(c.Request.Context(), id)
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

	news, err := h.news.Trx(request2.TxHandle(c)).Create(c.Request.Context(), dto.NewsCreate(cred))
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

	if err := h.news.Trx(request2.TxHandle(c)).Update(c.Request.Context(), id, dto.NewsUpdate(cred)); err != nil {
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

	err = h.news.Trx(request2.TxHandle(c)).Delete(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *NewsHandler) Export(c *gin.Context) ([]byte, error) {
	filter, err := request2.FilterJSON(c)
	if err != nil {
		return nil, err
	}

	f := meta.NewFilter(filter.OrderBy, filter.Where)

	cred := service.NewsExportInput{
		Filter: f,
	}

	return json.Marshal(cred)
}

func (h *NewsHandler) Import(c *gin.Context) ([]byte, error) {
	cred, errs, err := request2.BindImportFileRequest[request.NewsImportRequest](c, translator.EN.String())
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

func (h *NewsHandler) MassDelete(c *gin.Context) ([]byte, error) {
	cred, err := request2.BindJSON[request2.MultipleDeleteRequest](c)
	if err != nil {
		return nil, err
	}

	if cred == nil {
		return nil, apperr.ErrEmptyData
	}

	return json.Marshal(cred)
}
