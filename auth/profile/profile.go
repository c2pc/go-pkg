package profile

import (
	"context"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ErrNotFoundTranslate = translator.Translate{translator.RU: "Профиль не найден", translator.EN: "Profile not found"}
	ErrExistsTranslate   = translator.Translate{translator.RU: "Профиль уже зарегистрирован", translator.EN: "A profile is already registered"}
	ErrNotFound          = apperr.New("profile_not_found", apperr.WithTextTranslate(ErrNotFoundTranslate), apperr.WithCode(code.NotFound))
	ErrExists            = apperr.New("profile_exists_error", apperr.WithTextTranslate(ErrExistsTranslate), apperr.WithCode(code.InvalidArgument))
)

type IModel interface {
	GetUserId() int
}

type Profile[Model IModel, CreateInput, UpdateInput, UpdateProfileInput any] struct {
	Service     IProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	Request     IRequest[CreateInput, UpdateInput, UpdateProfileInput]
	Transformer ITransformer[Model]
}

type IProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput any] interface {
	Trx(db *gorm.DB) IProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	GetById(ctx context.Context, userID int) (*Model, error)
	GetByIds(ctx context.Context, userIDs ...int) ([]Model, error)
	Create(ctx context.Context, userID int, input CreateInput) (*Model, error)
	Update(ctx context.Context, userID int, input UpdateInput) error
	UpdateProfile(ctx context.Context, userID int, input UpdateProfileInput) error
	Delete(ctx context.Context, userID int) error
}

type IRequest[CreateInput, UpdateInput, UpdateProfileInput any] interface {
	CreateRequest(c *gin.Context) (*CreateInput, error)
	UpdateRequest(c *gin.Context) (*UpdateInput, error)
	UpdateProfileRequest(c *gin.Context) (*UpdateProfileInput, error)
}

type ITransformer[Model any] interface {
	Transform(m *Model) interface{}
	TransformList(models []Model) []interface{}
	TransformProfile(m *Model) interface{}
}
