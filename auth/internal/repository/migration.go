package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var MigrationSearchable = clause.FieldSearchable{}

var MigrationOrderBy = clause.FieldOrderBy{}

type IMigrationRepository interface {
	repository.Repository[IMigrationRepository, model.Migration]
}

type MigrationRepository struct {
	repository.Repo[model.Migration]
}

func NewMigrationRepository(db *gorm.DB) MigrationRepository {
	return MigrationRepository{
		Repo: repository.NewRepository[model.Migration](db, MigrationSearchable, MigrationOrderBy),
	}
}

func (r MigrationRepository) Trx(db *gorm.DB) IMigrationRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r MigrationRepository) With(models ...string) IMigrationRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r MigrationRepository) Joins(models ...string) IMigrationRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r MigrationRepository) Omit(columns ...string) IMigrationRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
