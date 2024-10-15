package db

import (
	"github.com/c2pc/go-pkg/v2/utils/level"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func ConnectMysql(url, debug string, maxIdleConn, maxOpenConn int) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(url), &gorm.Config{
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
