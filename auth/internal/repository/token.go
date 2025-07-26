package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var TokenSearchable = clause.FieldSearchable{
	"id":         {Column: `auth_tokens."id"`, Type: clause.Int},
	"user_id":    {Column: `auth_tokens."user_id"`, Type: clause.Int},
	"logged_at":  {Column: `auth_tokens."logged_at"`, Type: clause.DateTime},
	"updated_at": {Column: `auth_tokens."updated_at"`, Type: clause.DateTime},

	"user.id":          {Column: `"User"."id"`, Type: clause.Int, Join: "User"},
	"user.login":       {Column: `"User"."login"`, Type: clause.String, Join: "User"},
	"user.first_name":  {Column: `"User"."first_name"`, Type: clause.String, Join: "User"},
	"user.second_name": {Column: `"User"."second_name"`, Type: clause.String, Join: "User"},
	"user.last_name":   {Column: `"User"."last_name"`, Type: clause.String, Join: "User"},
}

var TokenOrderBy = clause.FieldOrderBy{
	"id":         {Column: `auth_tokens."id"`},
	"user_id":    {Column: `auth_tokens."user_id"`},
	"logged_at":  {Column: `auth_tokens."logged_at"`},
	"updated_at": {Column: `auth_tokens."updated_at"`},

	"user.id":          {Column: `"User"."id"`, Join: "User"},
	"user.login":       {Column: `"User"."login"`, Join: "User"},
	"user.first_name":  {Column: `"User"."first_name"`, Join: "User"},
	"user.second_name": {Column: `"User"."second_name"`, Join: "User"},
	"user.last_name":   {Column: `"User"."last_name"`, Join: "User"},
}

type ITokenRepository interface {
	repository.Repository[ITokenRepository, model.RefreshToken]
}

type TokenRepository struct {
	repository.Repo[model.RefreshToken]
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return TokenRepository{
		Repo: repository.NewRepository[model.RefreshToken](db, TokenSearchable, TokenOrderBy),
	}
}

func (r TokenRepository) Trx(db *gorm.DB) ITokenRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}
