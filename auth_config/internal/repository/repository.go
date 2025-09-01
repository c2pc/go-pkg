package repository

import (
	"gorm.io/gorm"
)

type Repositories struct {
	AuthConfigRepository AuthConfigRepository
}

func NewRepositories(db *gorm.DB) Repositories {
	return Repositories{
		AuthConfigRepository: NewAuthConfigRepository(db),
	}
}
