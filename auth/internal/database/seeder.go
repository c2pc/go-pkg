package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	seeders2 "github.com/c2pc/go-pkg/v2/auth/internal/database/seeders"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"gorm.io/gorm"
)

func SeedersRun(ctx context.Context, db *gorm.DB, repositories repository.Repositories, permissions []string) (admin *model.User, err error) {
	txHandle := db.Session(&gorm.Session{NewDB: true}).WithContext(ctx).Begin(&sql.TxOptions{})

	defer func() {
		if r := recover(); r != nil {
			txHandle.Rollback()
			err = errors.New(fmt.Sprint(r))
			return
		}
	}()

	admin, err = func() (*model.User, error) {
		perms, err := seeders2.PermissionSeeder(ctx, repositories.PermissionRepository.Trx(txHandle), permissions)
		if err != nil {
			return nil, err
		}

		role, err := seeders2.RoleSeeder(ctx, repositories.RoleRepository.Trx(txHandle), repositories.RolePermissionRepository.Trx(txHandle), perms)
		if err != nil {
			return nil, err
		}

		admin, err := seeders2.UserSeeder(ctx, repositories.UserRepository.Trx(txHandle), repositories.UserRoleRepository.Trx(txHandle), role.ID)
		if err != nil {
			return nil, err
		}

		return admin, nil
	}()
	if err != nil {
		txHandle.Rollback()
		return nil, err
	}

	txHandle.Commit()
	return admin, nil
}
