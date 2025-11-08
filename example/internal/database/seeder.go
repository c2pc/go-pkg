package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/c2pc/go-pkg/v2/example/internal/database/seeders"
	"github.com/c2pc/go-pkg/v2/example/profile"
	"gorm.io/gorm"
)

func SeedersRun(ctx context.Context, db *gorm.DB, profileRepository profile.IRepository, adminID int64) (err error) {
	txHandle := db.Session(&gorm.Session{NewDB: true}).WithContext(ctx).Begin(&sql.TxOptions{})

	defer func() {
		if r := recover(); r != nil {
			txHandle.Rollback()
			err = errors.New(fmt.Sprint(r))
			return
		}
	}()

	err = func() error {
		_, err := seeders.ProfileSeeder(ctx, profileRepository.Trx(txHandle), adminID)
		if err != nil {
			return err
		}

		return nil
	}()
	if err != nil {
		txHandle.Rollback()
		return err
	}

	txHandle.Commit()
	return nil
}
