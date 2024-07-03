package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var UserSearchable = clause.FieldSearchable{}

type IUserRepository interface {
	repository.Repository[IUserRepository, model.User]
}

type UserRepository struct {
	repository.Repo[model.User]
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return UserRepository{
		Repo: repository.NewRepository[model.User](db, UserSearchable),
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

func (r UserRepository) OrderBy(orderBy map[string]string) IUserRepository {
	r.Repo = r.Repo.OrderBy(orderBy)
	return r
}

func (r UserRepository) Omit(columns ...string) IUserRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
