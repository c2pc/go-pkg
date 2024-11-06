package db

import "gorm.io/gorm"

var gormConfig = &gorm.Config{
	PrepareStmt:            true,
	SkipDefaultTransaction: true,
}
