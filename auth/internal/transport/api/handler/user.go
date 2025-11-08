package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/dto"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService        service.IUserService
	tr                 mw.ITransaction
	profileTransformer profile.ITransformer
	profileRequest     profile.IRequest
}

func NewUserHandlers(
	userService service.IUserService,
	tr mw.ITransaction,
	profileTransformer profile.ITransformer,
	profileRequest profile.IRequest,
) *UserHandler {
	return &UserHandler{
		userService,
		tr,
		profileTransformer,
		profileRequest,
	}
}

func (h *UserHandler) Init(api *gin.RouterGroup) {
	user := api.Group("users")
	{
		user.GET("", h.List)
		user.GET("/:id", h.GetById)
		user.POST("", h.tr.DBTransaction, h.Create)
		user.PATCH("/:id", h.tr.DBTransaction, h.Update)
		user.DELETE("/:id", h.tr.DBTransaction, h.Delete)
	}
}

func (h *UserHandler) List(c *gin.Context) {
	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := meta.NewMeta(
		meta.NewPagination[model.User](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		meta.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.userService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.UserListTransform(c, m.Pagination, h.profileTransformer))
}

func (h *UserHandler) GetById(c *gin.Context) {
	id, err := request2.Id64(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.userService.GetById(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.UserTransform(data, h.profileTransformer))
}

func (h *UserHandler) Create(c *gin.Context) {
	c.Request = mcontext.WithOpActionRequest(c.Request, "Создание учетной записи")

	cred, err := request2.BindJSON[request.UserCreateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Request = mcontext.WithOpActionRequest(c.Request, "Создание учетной записи: "+cred.Login)

	var profileCred any
	if h.profileRequest != nil {
		profileCred, err = h.profileRequest.CreateRequest(c)
		if err != nil {
			response.Response(c, err)
			return
		}
	}

	user, err := h.userService.Trx(request2.TxHandle(c)).Create(c.Request.Context(), dto.UserCreate(cred), profileCred)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusCreated, transformer.UserTransform(user, h.profileTransformer))
}

func (h *UserHandler) Update(c *gin.Context) {
	c.Request = mcontext.WithOpActionRequest(c.Request, "Изменение учетной записи")
	id, err := request2.Id64(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	cred, err := request2.BindJSON[request.UserUpdateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	var profileCred any
	if h.profileRequest != nil {
		profileCred, err = h.profileRequest.UpdateRequest(c)
		if err != nil {
			response.Response(c, err)
			return
		}
	}

	userLogin, err := h.userService.Trx(request2.TxHandle(c)).Update(c.Request.Context(), id, dto.UserUpdate(cred), profileCred)
	if userLogin != "" {
		c.Request = mcontext.WithOpActionRequest(c.Request, "Изменение учетной записи: "+userLogin)
	}
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *UserHandler) Delete(c *gin.Context) {
	c.Request = mcontext.WithOpActionRequest(c.Request, "Удаление учетной записи")
	id, err := request2.Id64(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	userLogin, err := h.userService.Trx(request2.TxHandle(c)).Delete(c.Request.Context(), id)
	if userLogin != "" {
		c.Request = mcontext.WithOpActionRequest(c.Request, "Удаление учетной записи: "+userLogin)
	}
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}
