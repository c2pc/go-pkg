package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	"github.com/gin-gonic/gin"
)

type VersionHandler struct {
	versionService service.IVersionService
}

func NewVersionHandlers(
	versionService service.IVersionService,
) *VersionHandler {
	return &VersionHandler{
		versionService,
	}
}

func (h *VersionHandler) Init(api *gin.RouterGroup) {
	version := api.Group("versions")
	{
		version.GET("", h.Get)
	}
}

func (h *VersionHandler) Get(c *gin.Context) {
	data := h.versionService.Get(c.Request.Context())
	c.JSON(http.StatusOK, transformer.VersionTransform(data))
}
