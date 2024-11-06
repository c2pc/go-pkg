package repository

import (
	"gorm.io/gorm"
)

type Repositories struct {
	NewsRepository INewsRepository
}

func NewRepositories(db *gorm.DB) Repositories {
	return Repositories{
		NewsRepository: NewNewsRepository(db),
	}
}
