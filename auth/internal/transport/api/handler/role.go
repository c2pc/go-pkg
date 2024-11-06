package handler

import (
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/dto"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
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
		//role.POST("/mass-delete", h.DeleteMultiple)
		//role.POST("/mass-add", h.CreateMultiple)
		//role.POST("/mass-update", h.UpdateMultiple)
		role.GET("", h.List)
		role.GET("/:id", h.GetById)
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

	m := model2.NewMeta(
		model2.NewPagination[model.Role](cred.Limit, cred.Offset, cred.MustReturnTotalRows),
		model2.NewFilter(cred.OrderBy, cred.Where),
	)
	if err := h.roleService.List(c.Request.Context(), &m); err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusOK, transformer.RoleListTransform(c, m.Pagination))
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
	cred, err := request2.BindJSON[request.RoleCreateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	role, err := h.roleService.Trx(request2.TxHandle(c)).Create(c.Request.Context(), dto.RoleCreate(cred))
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusCreated, transformer.RoleTransform(role))
}

func (h *RoleHandler) Update(c *gin.Context) {
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

	if err := h.roleService.Trx(request2.TxHandle(c)).Update(c.Request.Context(), id, dto.RoleUpdate(cred)); err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := request2.Id(c)
	if err != nil {
		response.Response(c, err)
		return
	}

	err = h.roleService.Trx(request2.TxHandle(c)).Delete(c.Request.Context(), id)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

//func (h *RoleHandler) CreateMultiple(c *gin.Context) {
//	cred, err := request2.BindJSON[request2.MultipleAddRequest[request.RoleCreateRequest]](c)
//	if err != nil {
//		response.Response(c, err)
//		return
//	}
//
//	if cred == nil {
//		c.JSON(http.StatusOK, []int{})
//		return
//	}
//
//	multiple := model2.NewMultiple()
//	for _, input := range cred.Data {
//		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
//			err := v.StructCtx(c.Request.Context(), input)
//			if err == nil {
//				data, err := h.roleService.Create(c.Request.Context(), dto.RoleCreate(&input))
//				if err == nil {
//					multiple.AddID(data.ID)
//				}
//			}
//		}
//	}
//
//	c.JSON(http.StatusOK, multiple.IDs())
//}
//
//func (h *RoleHandler) UpdateMultiple(c *gin.Context) {
//	type UpdateRequest struct {
//		ID int `json:"id" binding:"required,gte=1"`
//		request.RoleUpdateRequest
//	}
//
//	cred, err := request2.BindJSON[request2.MultipleUpdateRequest[UpdateRequest]](c)
//	if err != nil {
//		response.Response(c, err)
//		return
//	}
//
//	if cred == nil {
//		c.JSON(http.StatusOK, []int{})
//		return
//	}
//
//	multiple := model2.NewMultiple()
//	for _, input := range cred.Data {
//		if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
//			err := v.StructCtx(c.Request.Context(), input)
//			if err == nil {
//				err = h.roleService.Update(c.Request.Context(), input.ID, dto.RoleUpdate(&input.RoleUpdateRequest))
//				if err == nil {
//					multiple.AddID(input.ID)
//				}
//			}
//		}
//
//	}
//
//	c.JSON(http.StatusOK, multiple.IDs())
//}
//
//func (h *RoleHandler) DeleteMultiple(c *gin.Context) {
//	cred, err := request2.BindJSON[request2.MultipleDeleteRequest](c)
//	if err != nil {
//		response.Response(c, err)
//		return
//	}
//
//	if cred == nil {
//		c.JSON(http.StatusOK, []int{})
//		return
//	}
//
//	multiple := model2.NewMultiple()
//	for _, id := range cred.Data {
//		if id > 0 {
//			err = h.roleService.Delete(c.Request.Context(), id)
//			if err == nil {
//				multiple.AddID(id)
//			}
//		}
//	}
//
//	c.JSON(http.StatusOK, multiple.IDs())
//}
