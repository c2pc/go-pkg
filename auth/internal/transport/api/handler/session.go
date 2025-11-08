package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/meta"
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
		session.DELETE("/:id", h.tr.DBTransaction, h.end)
	}
}
func (h *SessionHandler) list(c *gin.Context) {
	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := meta.NewMeta(
		meta.NewPagination[model.RefreshToken](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		meta.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.sessionService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.SessionListTransform(c, m.Pagination))
}

func (h *SessionHandler) end(c *gin.Context) {
	id, err := request2.Id64(c)
	if err != nil {
		c.Request = mcontext.WithOpActionRequest(c.Request, "Очистка сессии администратора")
		response.Response(c, err)
		return
	}

	userLogin, err := h.sessionService.Trx(request2.TxHandle(c)).End(c.Request.Context(), id)
	if err != nil {
		if userLogin != "" {
			c.Request = mcontext.WithOpActionRequest(c.Request, "Очистка сессии администратора: "+userLogin)
		}
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}
