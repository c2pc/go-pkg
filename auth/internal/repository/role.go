package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var RoleSearchable = clause.FieldSearchable{
	"id":   {`auth_roles."id"`, clause.Int, ""},
	"name": {`auth_roles."name"`, clause.String, ""},
}
var RoleOrderBy = clause.FieldOrderBy{
	"id":   {`auth_roles."id"`, ""},
	"name": {`auth_roles."name"`, ""},
}

type IRoleRepository interface {
	repository.Repository[IRoleRepository, model.Role]
}

type RoleRepository struct {
	repository.Repo[model.Role]
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return RoleRepository{
		Repo: repository.NewRepository[model.Role](db, RoleSearchable, RoleOrderBy),
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

func (r RoleRepository) Omit(columns ...string) IRoleRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
