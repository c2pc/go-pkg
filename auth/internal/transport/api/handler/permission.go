package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	permissionService service.IPermissionService
}

func NewPermissionHandlers(
	permissionService service.IPermissionService,

) *PermissionHandler {
	return &PermissionHandler{
		permissionService,
	}
}

func (h *PermissionHandler) Init(api *gin.RouterGroup) {
	permission := api.Group("permissions")
	{
		permission.GET("", h.List)
	}
}

func (h *PermissionHandler) List(c *gin.Context) {
	permissions, err := h.permissionService.List(c.Request.Context())
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.PermissionListTransform(c, permissions))
}
