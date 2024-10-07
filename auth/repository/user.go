package repository

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var userSearchable = clause.FieldSearchable{
	"id":          {`auth_users."id"`, clause.Int, ""},
	"login":       {`auth_users."login"`, clause.String, ""},
	"first_name":  {`auth_users."first_name"`, clause.String, ""},
	"second_name": {`auth_users."second_name"`, clause.String, ""},
	"last_name":   {`auth_users."last_name"`, clause.String, ""},
	"email":       {`auth_users."email"`, clause.String, ""},
	"phone":       {`auth_users."phone"`, clause.String, ""},
	"blocked":     {`auth_users."blocked"`, clause.Bool, ""},
}

var userOrderBy = clause.FieldOrderBy{
	"id":          {`auth_users."id"`, ""},
	"login":       {`auth_users."login"`, ""},
	"first_name":  {`auth_users."first_name"`, ""},
	"second_name": {`auth_users."second_name"`, ""},
	"last_name":   {`auth_users."last_name"`, ""},
	"email":       {`auth_users."email"`, ""},
	"phone":       {`auth_users."phone"`, ""},
	"blocked":     {`auth_users."blocked"`, ""},
}

type IUserRepository interface {
	repository.Repository[IUserRepository, model.User]
	GetUserWithPermissions(ctx context.Context, query string, args ...any) (*model.User, error)
}

type UserRepository struct {
	repository.Repo[model.User]
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return UserRepository{
		Repo: repository.NewRepository[model.User](db, userSearchable, userOrderBy),
	}
}

func (r UserRepository) Trx(db *gorm.DB) IUserRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r UserRepository) With(models ...string) IUserRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r UserRepository) Joins(models ...string) IUserRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r UserRepository) Omit(columns ...string) IUserRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}

func (r UserRepository) GetUserWithPermissions(ctx context.Context, query string, args ...any) (*model.User, error) {
	return r.Repo.
		With("roles", "roles.role_permissions", "roles.role_permissions.permission").
		Find(ctx, query, args...)
}
