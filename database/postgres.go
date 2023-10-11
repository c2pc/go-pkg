package database

import (
	"github.com/c2pc/go-pkg/logger"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"log"
	"time"
)

func ConnectPostgres(url, loggerID string, logInfo bool) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open("postgres://"+url), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if logInfo {
		logLevel := gormLogger.Info
		writer := logger.NewLogWriter(loggerID, true, 0)
		db.Logger = gormLogger.New(log.New(writer.Stdout, "\r\n\n", log.LstdFlags), gormLogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		})
	} else {
		db.Logger = gormLogger.Default.LogMode(gormLogger.Silent)
	}

	return db, nil
}
