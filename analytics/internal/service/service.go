package service

import (
	"context"

	"github.com/c2pc/go-pkg/v2/analytics/internal/models"
	"github.com/c2pc/go-pkg/v2/analytics/internal/repository"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
)

type AnalyticService interface {
	List(ctx context.Context, m *model2.Meta[models.Analytics]) error
	GetById(ctx context.Context, id int) (*models.Analytics, error)
}

type Analytic struct {
	analyticRepository repository.AnalyticsRepository
}

func NewAnalyticService(analyticRepository repository.AnalyticsRepository) AnalyticService {
	return Analytic{analyticRepository: analyticRepository}
}

func (a Analytic) List(ctx context.Context, m *model2.Meta[models.Analytics]) error {
	return a.analyticRepository.With("user").Omit("request_body", "response_body").Paginate(ctx, m, ``)
}

func (a Analytic) GetById(ctx context.Context, id int) (*models.Analytics, error) {
	data, err := a.analyticRepository.With("user").Find(ctx, "auth_analytics.id = ?", id)
	if err != nil {
		return nil, err
	}

	return data, nil
}
