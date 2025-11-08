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
	ErrExists            = apperr.New("profile_exists_error", apperr.WithTextTranslate(ErrExistsTranslate), apperr.WithCode(code.AlreadyExists))
)

type IModel interface {
	GetUserId() int64
}

type Profile struct {
	Service     IProfileService
	Request     IRequest
	Transformer ITransformer
}

type IProfileService interface {
	Trx(db *gorm.DB) IProfileService
	GetById(ctx context.Context, userID int64) (IModel, error)
	GetByIds(ctx context.Context, userIDs ...int64) ([]IModel, error)
	Create(ctx context.Context, userID int64, input any) (IModel, error)
	Update(ctx context.Context, userID int64, input any) error
	UpdateProfile(ctx context.Context, userID int64, input any) error
	Delete(ctx context.Context, userID int64) error
}

type IRequest interface {
	CreateRequest(c *gin.Context) (any, error)
	UpdateRequest(c *gin.Context) (any, error)
}

type ITransformer interface {
	Transform(m IModel) interface{}
	TransformList(models []IModel) []interface{}
	TransformProfile(m IModel) interface{}
}
