package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticService service.IAnalyticService
}

func NewAnalyticsHandler(analyticService service.IAnalyticService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticService: analyticService,
	}
}

func (h *AnalyticsHandler) Init(api *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	analytic := api.Group("/analytics", handlers...)
	{
		analytic.GET("/admins", h.GetListAdmin)
		analytic.GET("/admins/:id", h.GetByIdAdmin)
	}
}

func (h *AnalyticsHandler) GetListAdmin(c *gin.Context) {
	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := meta.NewMeta(
		meta.NewPagination[model.Analytic](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		meta.NewFilter(cred.OrderBy, cred.Where),
	)

	if err := h.analyticService.ListAdmin(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.AnalyticListTransform(c, m.Pagination))
}

func (h *AnalyticsHandler) GetByIdAdmin(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.analyticService.GetByIdAdmin(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.AnalyticTransform(data))
}
