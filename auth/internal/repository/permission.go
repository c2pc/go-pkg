package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var PermissionSearchable = clause.FieldSearchable{}
var PermissionOrderBy = clause.FieldOrderBy{}

type IPermissionRepository interface {
	repository.Repository[IPermissionRepository, model.Permission]
}

type PermissionRepository struct {
	repository.Repo[model.Permission]
}

func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return PermissionRepository{
		Repo: repository.NewRepository[model.Permission](db, PermissionSearchable, PermissionOrderBy),
	}
}

func (r PermissionRepository) Trx(db *gorm.DB) IPermissionRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}
