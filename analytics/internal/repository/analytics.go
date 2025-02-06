package repository

import (
	"github.com/c2pc/go-pkg/v2/analytics/internal/models"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var analyticSearchable = clause.FieldSearchable{
	"id":           {Column: `auth_analytics."id"`, Type: clause.Int},
	"user_id":      {Column: `auth_analytics."user_id"`, Type: clause.Int},
	"method":       {Column: `auth_analytics."method"`, Type: clause.String},
	"path":         {Column: `auth_analytics."path"`, Type: clause.String},
	"operation_id": {Column: `auth_analytics."operation_id"`, Type: clause.String},
	"status_code":  {Column: `auth_analytics."status_code"`, Type: clause.Int},
	"client_ip":    {Column: `auth_analytics."client_ip"`, Type: clause.String},
	"first_name":   {Column: `auth_analytics."first_name"`, Type: clause.String},
	"second_name":  {Column: `auth_analytics."second_name"`, Type: clause.String},
	"last_name":    {Column: `auth_analytics."last_name"`, Type: clause.String},
	"created_at":   {Column: `auth_analytics."created_at"`, Type: clause.DateTime},
}

var analyticOrderBy = clause.FieldOrderBy{
	"id":          {Column: `auth_analytics."id"`},
	"user_id":     {Column: `auth_analytics."user_id"`},
	"status_code": {Column: `auth_analytics."status_code"`},
	"client_ip":   {Column: `auth_analytics."client_ip"`},
	"first_name":  {Column: `auth_analytics."first_name"`},
	"second_name": {Column: `auth_analytics."second_name"`},
	"last_name":   {Column: `auth_analytics."last_name"`},
	"created_at":  {Column: `auth_analytics."created_at"`},
}

type AnalyticsRepository struct {
	repository.Repo[models.Analytics]
}

func NewAnalyticRepository(db *gorm.DB) AnalyticsRepository {
	return AnalyticsRepository{Repo: repository.NewRepository[models.Analytics](db, analyticSearchable, analyticOrderBy)}
}

func (r AnalyticsRepository) With(models ...string) AnalyticsRepository {
	r.Repo = r.Repo.With(models...)
	return r
}
