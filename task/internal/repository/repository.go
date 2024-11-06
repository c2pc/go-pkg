package repository

import (
	"gorm.io/gorm"
)

type Repositories struct {
	TaskRepository ITaskRepository
}

func NewRepositories(db *gorm.DB) Repositories {
	return Repositories{
		TaskRepository: NewTaskRepository(db),
	}
}
