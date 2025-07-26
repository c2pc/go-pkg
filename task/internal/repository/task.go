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

	"user.id":          {Column: `"User"."id"`, Type: clause.Int, Join: "User"},
	"user.login":       {Column: `"User"."login"`, Type: clause.Int, Join: "User"},
	"user.first_name":  {Column: `"User"."first_name"`, Type: clause.String, Join: "User"},
	"user.second_name": {Column: `"User"."second_name"`, Type: clause.String, Join: "User"},
	"user.last_name":   {Column: `"User"."last_name"`, Type: clause.String, Join: "User"},
}

var orderBy = clause.FieldOrderBy{
	"id":         {Column: `auth_tasks."id"`},
	"name":       {Column: `auth_tasks."name"`},
	"status":     {Column: `auth_tasks."status"`},
	"type":       {Column: `auth_tasks."type"`},
	"user_id":    {Column: `auth_tasks."user_id"`},
	"created_at": {Column: `auth_tasks."created_at"`},
	"updated_at": {Column: `auth_tasks."updated_at"`},

	"user.id":          {Column: `"User"."id"`, Join: "User"},
	"user.login":       {Column: `"User"."login"`, Join: "User"},
	"user.first_name":  {Column: `"User"."first_name"`, Join: "User"},
	"user.second_name": {Column: `"User"."second_name"`, Join: "User"},
	"user.last_name":   {Column: `"User"."last_name"`, Join: "User"},
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
