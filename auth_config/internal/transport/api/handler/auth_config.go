package handler

import (
	"encoding/json"
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth_config/internal/service"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/translator"

	"github.com/gin-gonic/gin"
)

var (
	ErrKeyRequired = apperr.New("key_is_required",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Ключ не найден", translator.EN: "Key not found"}),
		apperr.WithCode(code.InvalidArgument),
	)

	ErrKeyNotFound = apperr.New("key_not_found",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Ключ не найден", translator.EN: "Key not found"}),
		apperr.WithCode(code.NotFound),
	)
)

type AuthConfigHandler struct {
	authConfigService service.IAuthConfigService
	tr                mw.ITransaction
}

func NewAuthConfigHandlers(
	authConfigService service.IAuthConfigService,
	tr mw.ITransaction,
) *AuthConfigHandler {
	return &AuthConfigHandler{
		authConfigService,
		tr,
	}
}

func (h *AuthConfigHandler) GetService() service.IAuthConfigService {
	return h.authConfigService
}

func (h *AuthConfigHandler) Init(secured *gin.RouterGroup) {
	authConfig := secured.Group("configs")
	{
		authConfig.GET("", h.List)
		authConfig.PATCH("/:key", h.Update)
		authConfig.GET("/:key", h.GetByKey)
	}
}

func (h *AuthConfigHandler) List(c *gin.Context) {
	data, err := h.authConfigService.List(c.Request.Context())
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *AuthConfigHandler) GetByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.Response(c, ErrKeyRequired)
		return
	}

	data, err := h.authConfigService.GetByKey(c.Request.Context(), key)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, data)
}

func (h *AuthConfigHandler) Update(c *gin.Context) {
	c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Изменение конфигурации системы"))
	key := c.Param("key")
	if key == "" {
		response.Response(c, ErrKeyRequired)
		return
	}

	cred, err := request2.BindJSON[json.RawMessage](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	if cred == nil {
		response.Response(c, apperr.ErrEmptyData)
		return
	}

	action, err := h.authConfigService.Update(c.Request.Context(), key, *cred)
	if err != nil {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), action))
		response.Response(c, err)
		return
	}
	c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), action))

	c.Status(http.StatusOK)
}
