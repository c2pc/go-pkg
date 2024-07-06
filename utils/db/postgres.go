package db

import (
	"github.com/c2pc/go-pkg/v2/utils/level"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func ConnectPostgres(url, debug string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open("postgres://"+url), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	if level.Is(debug, level.DEVELOPMENT, level.TEST) {
		db.Logger = NewLogger(defaultLogger(debug))
	} else {
		db.Logger = gormLogger.Default.LogMode(gormLogger.Silent)
	}

	return db, nil
}
