package repository

import (
	"github.com/c2pc/go-pkg/v2/example/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var ProfileSearchable = clause.FieldSearchable{}

type IProfileRepository interface {
	repository.Repository[IProfileRepository, model.Profile]
}

type ProfileRepository struct {
	repository.Repo[model.Profile]
}

func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return ProfileRepository{
		Repo: repository.NewRepository[model.Profile](db, ProfileSearchable),
	}
}

func (r ProfileRepository) Trx(db *gorm.DB) IProfileRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r ProfileRepository) With(models ...string) IProfileRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r ProfileRepository) Joins(models ...string) IProfileRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r ProfileRepository) OrderBy(orderBy map[string]string) IProfileRepository {
	r.Repo = r.Repo.OrderBy(orderBy)
	return r
}

func (r ProfileRepository) Omit(columns ...string) IProfileRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
