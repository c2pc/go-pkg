package repository

import (
	"github.com/c2pc/go-pkg/v2/auth_config/internal/model"

	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"

	"gorm.io/gorm"
)

var AuthConfigSearchable = clause.FieldSearchable{
	"key": {Column: `auth_configs."key"`, Type: clause.String},
}

var AuthConfigOrderBy = clause.FieldOrderBy{
	"key": {Column: `auth_configs."key"`},
}

type AuthConfigRepository interface {
	repository.Repository[AuthConfigRepository, model.AuthConfig]
}

type AuthConfigRepositoryImpl struct {
	repository.Repo[model.AuthConfig]
}

func NewAuthConfigRepository(db *gorm.DB) AuthConfigRepositoryImpl {
	return AuthConfigRepositoryImpl{
		Repo: repository.NewRepository[model.AuthConfig](db, AuthConfigSearchable, AuthConfigOrderBy),
	}
}

func (r AuthConfigRepositoryImpl) Trx(db *gorm.DB) AuthConfigRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}
