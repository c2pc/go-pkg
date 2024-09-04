package database

import (
	"context"
	"database/sql"
	"github.com/c2pc/go-pkg/v2/example/database/seeders"
	"github.com/c2pc/go-pkg/v2/example/profile"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"gorm.io/gorm"
)

func SeedersRun(ctx context.Context, db *gorm.DB, profileRepository profile.IRepository, adminID int) error {
	txHandle := db.Session(&gorm.Session{NewDB: true}).WithContext(ctx).Begin(&sql.TxOptions{})

	defer func() {
		if r := recover(); r != nil {
			txHandle.Rollback()
			logger.Fatalf("%s - %v", apperr.ErrInternal, r)
			return
		}
	}()

	err := func() error {
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
