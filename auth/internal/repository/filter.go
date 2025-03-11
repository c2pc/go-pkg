package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var FilterSearchable = clause.FieldSearchable{
	"id":       {Column: `auth_filters."id"`, Type: clause.Int},
	"name":     {Column: `auth_filters."name"`, Type: clause.String},
	"endpoint": {Column: `auth_filters."endpoint"`, Type: clause.String},
}

var FilterOrderBy = clause.FieldOrderBy{
	"id":       {Column: `auth_filters."id"`},
	"name":     {Column: `auth_filters."name"`},
	"endpoint": {Column: `auth_filters."endpoint"`},
}

type IFilterRepository interface {
	repository.Repository[IFilterRepository, model.Filter]
}

type FilterRepository struct {
	repository.Repo[model.Filter]
}

func NewFilterRepository(db *gorm.DB) FilterRepository {
	return FilterRepository{
		Repo: repository.NewRepository[model.Filter](db, FilterSearchable, FilterOrderBy),
	}
}

func (r FilterRepository) Trx(db *gorm.DB) IFilterRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r FilterRepository) With(models ...string) IFilterRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r FilterRepository) Joins(models ...string) IFilterRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r FilterRepository) Omit(columns ...string) IFilterRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
