package repository

import (
	"github.com/c2pc/go-pkg/v2/task/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var searchable = clause.FieldSearchable{
	"id":         {Column: `auth_tasks."id"`, Type: clause.Int},
	"name":       {Column: `auth_tasks."name"`, Type: clause.String},
	"status":     {Column: `auth_tasks."status"`, Type: clause.String},
	"type":       {Column: `auth_tasks."type"`, Type: clause.String},
	"user_id":    {Column: `auth_tasks."user_id"`, Type: clause.Int},
	"created_at": {Column: `auth_tasks."created_at"`, Type: clause.DateTime},
	"updated_at": {Column: `auth_tasks."updated_at"`, Type: clause.DateTime},
}

var orderBy = clause.FieldOrderBy{
	"id":         {Column: `auth_tasks."id"`},
	"name":       {Column: `auth_tasks."name"`},
	"status":     {Column: `auth_tasks."status"`},
	"type":       {Column: `auth_tasks."type"`},
	"user_id":    {Column: `auth_tasks."user_id"`},
	"created_at": {Column: `auth_tasks."created_at"`},
	"updated_at": {Column: `auth_tasks."updated_at"`},
}

type ITaskRepository interface {
	repository.Repository[ITaskRepository, model.Task]
}

type TaskRepository struct {
	repository.Repo[model.Task]
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return TaskRepository{
		Repo: repository.NewRepository[model.Task](db, searchable, orderBy),
	}
}

func (r TaskRepository) Trx(db *gorm.DB) ITaskRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}

func (r TaskRepository) With(models ...string) ITaskRepository {
	r.Repo = r.Repo.With(models...)
	return r
}

func (r TaskRepository) Joins(models ...string) ITaskRepository {
	r.Repo = r.Repo.Joins(models...)
	return r
}

func (r TaskRepository) Omit(columns ...string) ITaskRepository {
	r.Repo = r.Repo.Omit(columns...)
	return r
}
