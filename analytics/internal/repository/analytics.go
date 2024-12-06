package repository

import (
	"github.com/c2pc/go-pkg/v2/analytics/internal/models"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/repository"
	"gorm.io/gorm"
)

var analyticSearchable = clause.FieldSearchable{
	"id":      {Column: `auth_analytics."id"`, Type: clause.Int},
	"user_id": {Column: `auth_analytics."user_id"`, Type: clause.String},
	"method":  {Column: `auth_analytics."method"`, Type: clause.String},
	"path":    {Column: `auth_analytics."path"`, Type: clause.String},
	"op_id":   {Column: `auth_analytics."operation_id"`, Type: clause.String},
}

var analyticOrderBy = clause.FieldOrderBy{
	"id":      {Column: `auth_analytics."id"`},
	"user_id": {Column: `auth_analytics."user_id"`},
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
