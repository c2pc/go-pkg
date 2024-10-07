package repository

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var UserRoleSearchable = clause.FieldSearchable{}
var UserRoleOrderBy = clause.FieldOrderBy{}

type IUserRoleRepository interface {
	repository.Repository[IUserRoleRepository, model.UserRole]
	GetUsersByRole(ctx context.Context, roleID int) ([]int, error)
}

type UserRoleRepository struct {
	repository.Repo[model.UserRole]
}

func NewUserRoleRepository(db *gorm.DB) UserRoleRepository {
	return UserRoleRepository{
		Repo: repository.NewRepository[model.UserRole](db, UserRoleSearchable, UserRoleOrderBy),
	}
}

func (r UserRoleRepository) Trx(db *gorm.DB) IUserRoleRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r UserRoleRepository) With(models ...string) IUserRoleRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r UserRoleRepository) Joins(models ...string) IUserRoleRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r UserRoleRepository) Omit(columns ...string) IUserRoleRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}

func (r UserRoleRepository) GetUsersByRole(ctx context.Context, roleID int) ([]int, error) {
	var result []int
	row := r.Model()

	res := r.Repo.DB().WithContext(ctx).Table(row.TableName()).Select("user_id").Where("role_id = ?", roleID).Scan(&result)
	if err := res.Error; err != nil {
		return nil, r.Error(err)
	}

	return result, nil
}
