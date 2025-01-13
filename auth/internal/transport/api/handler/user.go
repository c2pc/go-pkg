package handler

import (
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"net/http"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/transformer"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/mw"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
)

type UserHandler[Model, CreateInput, UpdateInput, UpdateProfileInput any] struct {
	userService        service.IUserService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	tr                 mw.ITransaction
	profileTransformer profile.ITransformer[Model]
	profileRequest     profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput]
}

func NewUserHandlers[Model, CreateInput, UpdateInput, UpdateProfileInput any](
	userService service.IUserService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	tr mw.ITransaction,
	profileTransformer profile.ITransformer[Model],
	profileRequest profile.IRequest[CreateInput, UpdateInput, UpdateProfileInput],
) *UserHandler[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	return &UserHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]{
		userService,
		tr,
		profileTransformer,
		profileRequest,
	}
}

func (h *UserHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) Init(api *gin.RouterGroup) {
	user := api.Group("users")
	{
		//user.POST("/mass-delete", h.DeleteMultiple)
		//user.POST("/mass-add", h.CreateMultiple)
		//user.POST("/mass-update", h.UpdateMultiple)
		user.GET("", h.List)
		user.GET("/:id", h.GetById)
		user.POST("", h.tr.DBTransaction, h.Create)
		user.PATCH("/:id", h.tr.DBTransaction, h.Update)
		user.DELETE("/:id", h.tr.DBTransaction, h.Delete)
	}
}

func (h *UserHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) List(c *gin.Context) {
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

	c.JSON(http.StatusOK, transformer.UserListTransform(c, m.Pagination, h.profileTransformer))
}

func (h *UserHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) GetById(c *gin.Context) {
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

	c.JSON(http.StatusOK, transformer.UserTransform(data, h.profileTransformer))
}

func (h *UserHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) Create(c *gin.Context) {
	cred, err := request2.BindJSON[request.UserCreateRequest](c)
	if err != nil {
		response.Response(c, err)
		return
	}

	var profileCred *CreateInput
	if h.profileRequest != nil {
		profileCred, err = h.profileRequest.CreateRequest(c)
		if err != nil {
			response.Response(c, err)
			return
		}
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
	}, profileCred)
	if err != nil {
		response.Response(c, err)
		return
	}

	c.JSON(http.StatusCreated, transformer.UserTransform(user, h.profileTransformer))
}

func (h *UserHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) Update(c *gin.Context) {
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

	var profileCred *UpdateInput
	if h.profileRequest != nil {
		profileCred, err = h.profileRequest.UpdateRequest(c)
		if err != nil {
			response.Response(c, err)
			return
		}
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
	}, profileCred); err != nil {
		response.Response(c, err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *UserHandler[Model, CreateInput, UpdateInput, UpdateProfileInput]) Delete(c *gin.Context) {
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

//func (h *UserHandler) CreateMultiple(c *gin.Context) {
//	cred, err := request2.BindJSON[request2.MultipleAddRequest[request.UserCreateRequest]](c)
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
//				data, err := h.userService.Create(c.Request.Context(), dto.UserCreate(&input))
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
//func (h *UserHandler) UpdateMultiple(c *gin.Context) {
//	type UpdateRequest struct {
//		ID int `json:"id" binding:"required,gte=1"`
//		request.UserUpdateRequest
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
//				err = h.userService.Update(c.Request.Context(), input.ID, dto.UserUpdate(&input.UserUpdateRequest))
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
//func (h *UserHandler) DeleteMultiple(c *gin.Context) {
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
//			err = h.userService.Delete(c.Request.Context(), id)
//			if err == nil {
//				multiple.AddID(id)
//			}
//		}
//	}
//
//	c.JSON(http.StatusOK, multiple.IDs())
//}
