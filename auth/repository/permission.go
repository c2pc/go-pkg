package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var PermissionSearchable = clause.FieldSearchable{}

type IPermissionRepository interface {
	repository.Repository[IPermissionRepository, model.Permission]
}

type PermissionRepository struct {
	repository.Repo[model.Permission]
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return PermissionRepository{
		Repo: repository.NewRepository[model.Permission](db, PermissionSearchable),
	}
}

func (r PermissionRepository) Trx(db *gorm.DB) IPermissionRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r PermissionRepository) With(models ...string) IPermissionRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r PermissionRepository) Joins(models ...string) IPermissionRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r PermissionRepository) OrderBy(orderBy map[string]string) IPermissionRepository {
	r.Repo = r.Repo.OrderBy(orderBy)
	return r
}

func (r PermissionRepository) Omit(columns ...string) IPermissionRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
