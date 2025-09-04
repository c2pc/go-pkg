package handler

import (
	"encoding/json"
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth_config/internal/service"
	"github.com/c2pc/go-pkg/v2/auth_config/transformer"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
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
	transformers      transformer.AuthConfigTransformers
}

func NewAuthConfigHandlers(
	authConfigService service.IAuthConfigService,
	tr mw.ITransaction,
	transformers transformer.AuthConfigTransformers,
) *AuthConfigHandler {
	return &AuthConfigHandler{
		authConfigService,
		tr,
		transformers,
	}
}

func (h *AuthConfigHandler) GetService() service.IAuthConfigService {
	return h.authConfigService
}

func (h *AuthConfigHandler) Init(secured *gin.RouterGroup, unsecured *gin.RouterGroup) {
	authConfig := secured.Group("auth-configs")
	{
		authConfig.GET("", h.List)
		authConfig.PATCH("/:key", h.Update)
		authConfig.GET("/:key", h.GetByKey)
	}
}

func (h *AuthConfigHandler) List(c *gin.Context) {
	var resp []any

	authConfigs, err := h.authConfigService.List(c.Request.Context())
	if err != nil {
		response.Response(c, err)
		return
	}

	for i := range authConfigs {
		tmpl, ok := h.transformers[authConfigs[i].Key]
		if !ok {
			continue
		}

		value, err := tmpl.Transform(authConfigs[i].Value)
		if err != nil {
			response.Response(c, err)
			return
		}
		resp = append(resp, value)
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthConfigHandler) GetByKey(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		response.Response(c, ErrKeyRequired)
		return
	}

	tmpl := h.transformers[key]
	if tmpl == nil {
		response.Response(c, ErrKeyNotFound)
		return
	}

	data, err := h.authConfigService.GetByKey(c.Request.Context(), key)
	if err != nil {
		response.Response(c, err)
		return
	}

	value, err := tmpl.Transform(data.Value)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, value)
}

func (h *AuthConfigHandler) Update(c *gin.Context) {
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

	tmpl := h.transformers[key]
	if tmpl == nil {
		response.Response(c, ErrKeyNotFound)
		return
	}

	err = tmpl.Check(*cred)
	if err != nil {
		response.Response(c, err)
		return
	}

	if err := h.authConfigService.Update(c.Request.Context(), key, *cred); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, nil)
}
