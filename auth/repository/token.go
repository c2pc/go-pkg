package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var TokenSearchable = clause.FieldSearchable{}
var TokenOrderBy = clause.FieldOrderBy{}

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

func (r TokenRepository) With(models ...string) ITokenRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r TokenRepository) Joins(models ...string) ITokenRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r TokenRepository) Omit(columns ...string) ITokenRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
