package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"

	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"

	"gorm.io/gorm"
)

var ConfigSearchable = clause.FieldSearchable{
	"key": {Column: `auth_configs."key"`, Type: clause.String},
}

var ConfigOrderBy = clause.FieldOrderBy{
	"key": {Column: `auth_configs."key"`},
}

type IConfigRepository interface {
	repository.Repository[IConfigRepository, model.Config]
}

type ConfigRepository struct {
	repository.Repo[model.Config]
}

func NewConfigRepository(db *gorm.DB) IConfigRepository {
	return ConfigRepository{
		Repo: repository.NewRepository[model.Config](db, ConfigSearchable, ConfigOrderBy),
	}
}

func (r ConfigRepository) Trx(db *gorm.DB) IConfigRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}
