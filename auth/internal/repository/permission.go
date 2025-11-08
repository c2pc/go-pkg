package repository

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var PermissionSearchable = clause.FieldSearchable{
	"name": {Column: `auth_permissions."name"`, Type: clause.String},
}
var PermissionOrderBy = clause.FieldOrderBy{
	"name": {Column: `auth_permissions."name"`},
}

type IPermissionRepository interface {
	repository.Repository[IPermissionRepository, model.Permission]
	GetList(ctx context.Context) ([]model.Permission, error)
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

func (r PermissionRepository) GetList(ctx context.Context) ([]model.Permission, error) {
	return r.Repo.List(ctx, &meta.Filter{
		OrderBy: []clause.ExpressionOrderBy{{Column: "name", Order: clause.OrderByAsc}},
	}, ``)
}
