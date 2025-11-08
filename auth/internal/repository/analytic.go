package repository

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var analyticAdminSearchable = clause.FieldSearchable{
	"id":           {Column: `id`, Type: clause.Int},
	"user_id":      {Column: `user_id`, Type: clause.Int},
	"method":       {Column: `method`, Type: clause.String},
	"path":         {Column: `path`, Type: clause.String},
	"operation_id": {Column: `operation_id`, Type: clause.String},
	"status_code":  {Column: `status_code`, Type: clause.Int},
	"client_ip":    {Column: `client_ip`, Type: clause.String},
	"name":         {Column: `name`, Type: clause.String},
	"created_at":   {Column: `created_at`, Type: clause.DateTime},
	"login":        {Column: `login`, Type: clause.String},
	"error":        {Column: `error`, Type: clause.String},
	"action":       {Column: `action`, Type: clause.String},
}

var analyticAdminOrderBy = clause.FieldOrderBy{
	"id":           {Column: `id`},
	"user_id":      {Column: `user_id`},
	"method":       {Column: `method`},
	"path":         {Column: `path`},
	"operation_id": {Column: `operation_id`},
	"status_code":  {Column: `status_code`},
	"client_ip":    {Column: `client_ip`},
	"name":         {Column: `name`},
	"created_at":   {Column: `created_at`},
	"login":        {Column: `login`},
	"error":        {Column: `error`},
	"action":       {Column: `action`},
}

type IAnalyticRepository interface {
	repository.Repository[IAnalyticRepository, model.Analytic]
}

type AnalyticRepository struct {
	repository.Repo[model.Analytic]
}

func NewAnalyticRepository(db *gorm.DB) IAnalyticRepository {
	return AnalyticRepository{
		Repo: repository.NewRepository[model.Analytic](db, analyticAdminSearchable, analyticAdminOrderBy),
	}
}

func (r AnalyticRepository) Trx(db *gorm.DB) IAnalyticRepository {
	r.Repo = r.Repo.Trx(db)
	return r
}
