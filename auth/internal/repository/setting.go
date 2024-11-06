package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var SettingSearchable = clause.FieldSearchable{}
var SettingOrderBy = clause.FieldOrderBy{}

type ISettingRepository interface {
	repository.Repository[ISettingRepository, model.Setting]
}

type SettingRepository struct {
	repository.Repo[model.Setting]
}

func NewSettingRepository(db *gorm.DB) SettingRepository {
	return SettingRepository{
		Repo: repository.NewRepository[model.Setting](db, SettingSearchable, SettingOrderBy),
	}
}

func (r SettingRepository) Trx(db *gorm.DB) ISettingRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r SettingRepository) With(models ...string) ISettingRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r SettingRepository) Joins(models ...string) ISettingRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r SettingRepository) Omit(columns ...string) ISettingRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
