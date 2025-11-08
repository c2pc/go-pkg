package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/dto"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"

	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	roleService service.IRoleService
	tr          mw.ITransaction
}

func NewRoleHandlers(
	roleService service.IRoleService,
	tr mw.ITransaction,
) *RoleHandler {
	return &RoleHandler{
		roleService,
		tr,
	}
}

func (h *RoleHandler) Init(api *gin.RouterGroup) {
	role := api.Group("roles")
	{
		role.GET("", h.List)
		role.GET("/:id", h.GetById)
		role.GET("/:id/users", h.UserList)
		role.POST("", h.tr.DBTransaction, h.Create)
		role.PATCH("/:id", h.tr.DBTransaction, h.Update)
		role.DELETE("/:id", h.tr.DBTransaction, h.Delete)
	}
}

func (h *RoleHandler) List(c *gin.Context) {
	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := meta.NewMeta(
		meta.NewPagination[model.Role](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		meta.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.roleService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.RoleListTransform(c, m.Pagination))
}

func (h *RoleHandler) UserList(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	cred, err := request2.Meta(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	m := meta.NewMeta(
		meta.NewPagination[model.UserRole](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		meta.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.roleService.UserList(c.Request.Context(), id, &m); err != nil {
		response.Response(c, err)
		return
	}

	userList := []model.User{}
	for _, row := range m.Pagination.Rows {
		userList = append(userList, *row.User)
	}

	c.JSON(http.StatusOK, transformer.UserRoleListTransform(c, m.Pagination))
}

func (h *RoleHandler) GetById(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	data, err := h.roleService.GetById(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.RoleTransform(data))
}

func (h *RoleHandler) Create(c *gin.Context) {
	c.Request = mcontext.WithOpActionRequest(c.Request, "Создание роли администраторов")

	cred, err := request2.BindJSON[request.RoleCreateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Request = mcontext.WithOpActionRequest(c.Request, "Создание роли администраторов: "+cred.Name)

	role, err := h.roleService.Trx(request2.TxHandle(c)).Create(c.Request.Context(), dto.RoleCreate(cred))
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusCreated, transformer.RoleTransform(role))
}

func (h *RoleHandler) Update(c *gin.Context) {
	c.Request = mcontext.WithOpActionRequest(c.Request, "Изменение роли администраторов")

	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	cred, err := request2.BindJSON[request.RoleUpdateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	roleName, err := h.roleService.Trx(request2.TxHandle(c)).Update(c.Request.Context(), id, dto.RoleUpdate(cred))
	if roleName != "" {
		c.Request = mcontext.WithOpActionRequest(c.Request, "Изменение роли администраторов: "+roleName)
	}
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	c.Request = mcontext.WithOpActionRequest(c.Request, "Удаление роли администраторов")

	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	roleName, err := h.roleService.Trx(request2.TxHandle(c)).Delete(c.Request.Context(), id)
	if roleName != "" {
		c.Request = mcontext.WithOpActionRequest(c.Request, "Удаление роли администраторов: "+roleName)
	}
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}
