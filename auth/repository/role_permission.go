package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var RolePermissionSearchable = clause.FieldSearchable{}
var RolePermissionOrderBy = clause.FieldOrderBy{}

type IRolePermissionRepository interface {
	repository.Repository[IRolePermissionRepository, model.RolePermission]
}

type RolePermissionRepository struct {
	repository.Repo[model.RolePermission]
}

func NewRolePermissionRepository(db *gorm.DB) RolePermissionRepository {
	return RolePermissionRepository{
		Repo: repository.NewRepository[model.RolePermission](db, RolePermissionSearchable, RolePermissionOrderBy),
	}
}

func (r RolePermissionRepository) Trx(db *gorm.DB) IRolePermissionRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r RolePermissionRepository) With(models ...string) IRolePermissionRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r RolePermissionRepository) Joins(models ...string) IRolePermissionRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r RolePermissionRepository) Omit(columns ...string) IRolePermissionRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
