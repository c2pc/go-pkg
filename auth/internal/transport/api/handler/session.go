package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	sessionService service.ISessionService
	tr             mw.ITransaction
}

func NewSessionHandlers(
	sessionService service.ISessionService,
	tr mw.ITransaction,
) *SessionHandler {
	return &SessionHandler{
		sessionService,
		tr,
	}
}

func (h *SessionHandler) Init(api *gin.RouterGroup) {
	session := api.Group("/sessions")
	{
		session.GET("", h.list)
		session.POST("/:id/end", h.tr.DBTransaction, h.end)
	}
}
func (h *SessionHandler) list(c *gin.Context) {
	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := model2.NewMeta(
		model2.NewPagination[model.RefreshToken](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		model2.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.sessionService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.SessionListTransform(c, m.Pagination))
}

func (h *SessionHandler) end(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	if err := h.sessionService.Trx(request2.TxHandle(c)).End(c.Request.Context(), id); err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}
