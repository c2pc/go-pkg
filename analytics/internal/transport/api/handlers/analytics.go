package handlers

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/analytics/internal/models"
	"github.com/c2pc/go-pkg/v2/analytics/internal/service"
	"github.com/c2pc/go-pkg/v2/analytics/internal/transport/api/transformer"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticsService service.AnalyticService
}

func NewAnalyticsHandler(analyticsService service.AnalyticService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

func (h *AnalyticsHandler) Init(api *gin.RouterGroup, handlers ...gin.HandlerFunc) {
	analytics := api.Group("/query-history", handlers...)
	{
		analytics.GET("", h.GetList)
		analytics.GET("/:id", h.GetById)
	}
}

func (h *AnalyticsHandler) GetList(c *gin.Context) {
	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := model2.NewMeta(
		model2.NewPagination[models.Analytics](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		model2.NewFilter(cred.OrderBy, cred.Where),
	)

	if err := h.analyticsService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.AnalyticListTransform(c, m.Pagination))
}

func (h *AnalyticsHandler) GetById(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.analyticsService.GetById(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.AnalyticTransform(data))
}
