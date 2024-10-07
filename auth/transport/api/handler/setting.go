package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/service"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type SettingHandler struct {
	settingService service.ISettingService
	tr             mw.ITransaction
}

func NewSettingHandlers(
	settingService service.ISettingService,
	tr mw.ITransaction,
) *SettingHandler {
	return &SettingHandler{
		settingService,
		tr,
	}
}

func (h *SettingHandler) Init(api *gin.RouterGroup) {
	setting := api.Group("settings")
	{
		setting.GET("", h.Get)
		setting.PATCH("", h.tr.DBTransaction, h.Update)
	}
}

func (h *SettingHandler) Get(c *gin.Context) {
	data, err := h.settingService.Get(c.Request.Context())
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.SettingTransform(data))
}

func (h *SettingHandler) Update(c *gin.Context) {
	cred, err := request2.BindJSON[request.SettingUpdateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	if err := h.settingService.Trx(request2.TxHandle(c)).Update(c.Request.Context(), service.SettingUpdateInput{
		Settings: cred.Settings,
	}); err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}
