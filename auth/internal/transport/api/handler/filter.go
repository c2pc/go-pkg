package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type FilterHandler struct {
	filterService service.IFilterService
	tr            mw.ITransaction
}

func NewFilterHandlers(
	filterService service.IFilterService,
	tr mw.ITransaction,
) *FilterHandler {
	return &FilterHandler{
		filterService,
		tr,
	}
}

func (h *FilterHandler) Init(api *gin.RouterGroup) {
	filter := api.Group("filters")
	{
		filter.GET("", h.List)
		filter.GET("/:id", h.GetById)
		filter.POST("", h.tr.DBTransaction, h.Create)
		filter.PATCH("/:id", h.tr.DBTransaction, h.Update)
		filter.DELETE("/:id", h.tr.DBTransaction, h.Delete)
	}
}

func (h *FilterHandler) List(c *gin.Context) {
	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := model2.NewMeta(
		model2.NewPagination[model.Filter](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		model2.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.filterService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.FilterListTransform(c, m.Pagination))
}

func (h *FilterHandler) GetById(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.filterService.GetById(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.FilterTransform(data))
}

func (h *FilterHandler) Create(c *gin.Context) {
	cred, err := request2.BindJSON[request.FilterCreateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	filter, err := h.filterService.Trx(request2.TxHandle(c)).Create(c.Request.Context(), service.FilterCreateInput{
		Name:     cred.Name,
		Endpoint: cred.Endpoint,
		Value:    cred.Value,
	})
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusCreated, transformer.FilterTransform(filter))
}

func (h *FilterHandler) Update(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	cred, err := request2.BindJSON[request.FilterUpdateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	if err := h.filterService.Trx(request2.TxHandle(c)).Update(c.Request.Context(), id, service.FilterUpdateInput{
		Name:  cred.Name,
		Value: cred.Value,
	}); err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *FilterHandler) Delete(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	err = h.filterService.Trx(request2.TxHandle(c)).Delete(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}
