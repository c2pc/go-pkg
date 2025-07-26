package profile

import (
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var Searchable = clause.FieldSearchable{}
var OrderBy = clause.FieldOrderBy{}

type IRepository interface {
	repository.Repository[IRepository, Profile]
}

type Repository struct {
	repository.Repo[Profile]
}

func NewProfileRepository(db *gorm.DB) Repository {
	return Repository{
		Repo: repository.NewRepository[Profile](db, Searchable, OrderBy),
	}
}

func (r Repository) Trx(db *gorm.DB) IRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}
