package repository

import (
	"github.com/c2pc/go-pkg/v2/task/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var searchable = clause.FieldSearchable{
	"id":         {`auth_tasks."id"`, clause.Int, ""},
	"name":       {`auth_tasks."name"`, clause.String, ""},
	"status":     {`auth_tasks."status"`, clause.String, ""},
	"created_at": {`auth_tasks."created_at"`, clause.DateTime, ""},
	"updated_at": {`auth_tasks."updated_at"`, clause.DateTime, ""},
}

var orderBy = clause.FieldOrderBy{
	"id":         {`auth_tasks."id"`, ""},
	"name":       {`auth_tasks."name"`, ""},
	"status":     {`auth_tasks."status"`, ""},
	"created_at": {`auth_tasks."created_at"`, ""},
	"updated_at": {`auth_tasks."updated_at"`, ""},
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
