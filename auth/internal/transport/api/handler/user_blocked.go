package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
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

type UserBlockedHandler struct {
	usersBlockedService service.IUserBlockedService
	tr                  mw.ITransaction
	limiter             *fx.LimiterHolder
}

func NewUserBlockedHandlers(
	usersBlockedService service.IUserBlockedService,
	tr mw.ITransaction,
	limiter *fx.LimiterHolder,
) *UserBlockedHandler {
	return &UserBlockedHandler{
		usersBlockedService,
		tr,
		limiter,
	}
}

func (h *UserBlockedHandler) Init(api *gin.RouterGroup) {
	usersBlocked := api.Group("/users-blocked")
	{
		usersBlocked.GET("", h.list)
		usersBlocked.DELETE("/:id", h.tr.DBTransaction, h.unlock)
	}
}
func (h *UserBlockedHandler) list(c *gin.Context) {
	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := meta.NewMeta(
		meta.NewPagination[model.UserBlocked](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		meta.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.usersBlockedService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.UserBlockedListTransform(c, h.limiter, m.Pagination))
}

func (h *UserBlockedHandler) unlock(c *gin.Context) {
	id, err := request2.Id64(c)
	if err != nil {
		c.Request = mcontext.WithOpActionRequest(c.Request, "Очистка сессии администратора")
		response.Response(c, err)
		return
	}

	userLogin, err := h.usersBlockedService.Trx(request2.TxHandle(c)).Unlock(c.Request.Context(), id)
	if err != nil {
		if userLogin != "" {
			c.Request = mcontext.WithOpActionRequest(c.Request, "Очистка сессии администратора: "+userLogin)
		}
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}
