package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"

	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"

	"gorm.io/gorm"
)

var ConfigFileSearchable = clause.FieldSearchable{}

var ConfigFileOrderBy = clause.FieldOrderBy{}

type IConfigFileRepository interface {
	repository.Repository[IConfigFileRepository, model.ConfigFile]
}

type ConfigFileRepository struct {
	repository.Repo[model.ConfigFile]
}

func NewConfigFileRepository(db *gorm.DB) IConfigFileRepository {
	return ConfigFileRepository{
		Repo: repository.NewRepository[model.ConfigFile](db, ConfigFileSearchable, ConfigFileOrderBy),
	}
}

func (r ConfigFileRepository) Trx(db *gorm.DB) IConfigFileRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}
