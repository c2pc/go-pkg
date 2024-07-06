package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var RoleSearchable = clause.FieldSearchable{}

type IRoleRepository interface {
	repository.Repository[IRoleRepository, model.Role]
}

type RoleRepository struct {
	repository.Repo[model.Role]
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return RoleRepository{
		Repo: repository.NewRepository[model.Role](db, RoleSearchable),
	}
}

func (r RoleRepository) Trx(db *gorm.DB) IRoleRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r RoleRepository) With(models ...string) IRoleRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r RoleRepository) Joins(models ...string) IRoleRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r RoleRepository) OrderBy(orderBy map[string]string) IRoleRepository {
	r.Repo = r.Repo.OrderBy(orderBy)
	return r
}

func (r RoleRepository) Omit(columns ...string) IRoleRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
