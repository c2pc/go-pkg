package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var userBlockedSearchable = clause.FieldSearchable{
	"id":          {Column: `"User"."id"`, Type: clause.Int, Join: "User"},
	"login":       {Column: `"User"."login"`, Type: clause.String, Join: "User"},
	"first_name":  {Column: `"User"."first_name"`, Type: clause.String, Join: "User"},
	"second_name": {Column: `"User"."second_name"`, Type: clause.String, Join: "User"},
	"last_name":   {Column: `"User"."last_name"`, Type: clause.String, Join: "User"},
	"is_domain":   {Column: `"User"."is_domain"`, Type: clause.Bool, Join: "User"},
	"blocked_at":  {Column: `auth_users_blocked."blocked_at"`, Type: clause.DateTime},
}

var userBlockedOrderBy = clause.FieldOrderBy{
	"id":          {Column: `"User"."id"`, Join: "User"},
	"login":       {Column: `"User"."login"`, Join: "User"},
	"first_name":  {Column: `"User"."first_name"`, Join: "User"},
	"second_name": {Column: `"User"."second_name"`, Join: "User"},
	"last_name":   {Column: `"User"."last_name"`, Join: "User"},
	"is_domain":   {Column: `"User"."is_domain"`, Join: "User"},
	"blocked_at":  {Column: `auth_users_blocked."blocked_at"`},
}

type IUserBlockedRepository interface {
	repository.Repository[IUserBlockedRepository, model.UserBlocked]
}

type UserBlockedRepository struct {
	repository.Repo[model.UserBlocked]
}

func NewUserBlockedRepository(db *gorm.DB) UserBlockedRepository {
	return UserBlockedRepository{
		Repo: repository.NewRepository[model.UserBlocked](db, userBlockedSearchable, userBlockedOrderBy),
	}
}

func (r UserBlockedRepository) Trx(db *gorm.DB) IUserBlockedRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}
