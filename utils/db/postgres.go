package db

import (
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func ConnectPostgres(url string, maxIdleConn, maxOpenConn int) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open("postgres://"+url), gormConfig)
	if err != nil {
		return nil, err
	}

	db.Logger = NewLogger(gormLogger.Config{
		SlowThreshold:             1 * time.Second,
		Colorful:                  false,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
		LogLevel:                  gormLogger.Info,
	})

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if maxIdleConn > 0 {
		sqlDB.SetMaxIdleConns(maxIdleConn)
	} else {
		sqlDB.SetMaxIdleConns(10)
	}

	if maxOpenConn > 0 {
		sqlDB.SetMaxOpenConns(maxOpenConn)
	} else {
		sqlDB.SetMaxOpenConns(100)
	}

	return db, nil
}
