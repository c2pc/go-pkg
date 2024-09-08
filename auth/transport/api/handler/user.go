package handler

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/service"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/transport/api/transformer"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler struct {
	userService service.IUserService
	tr          mw.ITransaction
}

func NewUserHandlers(
	userService service.IUserService,
	tr mw.ITransaction,
) *UserHandler {
	return &UserHandler{
		userService,
		tr,
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

	m := model2.NewMeta(
		model2.NewPagination[model.User](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		model2.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.userService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.UserListTransform(c, m.Pagination))
}

func (h *UserHandler) GetById(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.userService.GetById(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.UserTransform(data))
}

func (h *UserHandler) Create(c *gin.Context) {
	cred, err := request2.BindJSON[request.UserCreateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	user, err := h.userService.Trx(request2.TxHandle(c)).Create(c.Request.Context(), service.UserCreateInput{
		Login:      cred.Login,
		FirstName:  cred.FirstName,
		SecondName: cred.SecondName,
		LastName:   cred.LastName,
		Password:   cred.Password,
		Email:      cred.Email,
		Phone:      cred.Phone,
		Roles:      cred.Roles,
	})
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusCreated, transformer.UserTransform(user))
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	cred, err := request2.BindJSON[request.UserUpdateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	if err := h.userService.Trx(request2.TxHandle(c)).Update(c.Request.Context(), id, service.UserUpdateInput{
		Login:      cred.Login,
		FirstName:  cred.FirstName,
		SecondName: cred.SecondName,
		LastName:   cred.LastName,
		Password:   cred.Password,
		Email:      cred.Email,
		Phone:      cred.Phone,
		Roles:      cred.Roles,
	}); err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	err = h.userService.Trx(request2.TxHandle(c)).Delete(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}
