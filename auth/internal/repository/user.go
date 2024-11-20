package repository

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var userSearchable = clause.FieldSearchable{
	"id":          {Column: `auth_users."id"`, Type: clause.Int},
	"login":       {Column: `auth_users."login"`, Type: clause.String},
	"first_name":  {Column: `auth_users."first_name"`, Type: clause.String},
	"second_name": {Column: `auth_users."second_name"`, Type: clause.String},
	"last_name":   {Column: `auth_users."last_name"`, Type: clause.String},
	"email":       {Column: `auth_users."email"`, Type: clause.String},
	"phone":       {Column: `auth_users."phone"`, Type: clause.String},
	"blocked":     {Column: `auth_users."blocked"`, Type: clause.Bool},
}

var userOrderBy = clause.FieldOrderBy{
	"id":          {Column: `auth_users."id"`},
	"login":       {Column: `auth_users."login"`},
	"first_name":  {Column: `auth_users."first_name"`},
	"second_name": {Column: `auth_users."second_name"`},
	"last_name":   {Column: `auth_users."last_name"`},
	"email":       {Column: `auth_users."email"`},
	"phone":       {Column: `auth_users."phone"`},
	"blocked":     {Column: `auth_users."blocked"`},
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
