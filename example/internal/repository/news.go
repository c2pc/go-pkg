package repository

import (
	"github.com/c2pc/go-pkg/v2/example/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var newsSearchable = clause.FieldSearchable{
	"id":    {Column: `news."id"`, Type: clause.Int},
	"title": {Column: `news."title"`, Type: clause.String},
}

var newsOrderBy = clause.FieldOrderBy{
	"id":    {Column: `news."id"`},
	"title": {Column: `news."title"`},
}

type INewsRepository interface {
	repository.Repository[INewsRepository, model.News]
}

type NewsRepository struct {
	repository.Repo[model.News]
}

func NewNewsRepository(db *gorm.DB) NewsRepository {
	return NewsRepository{
		Repo: repository.NewRepository[model.News](db, newsSearchable, newsOrderBy),
	}
}

func (r NewsRepository) Trx(db *gorm.DB) INewsRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r NewsRepository) With(models ...string) INewsRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r NewsRepository) Joins(models ...string) INewsRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r NewsRepository) Omit(columns ...string) INewsRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
