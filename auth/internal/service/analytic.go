package service

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/meta"
)

type IAnalyticService interface {
	ListAdmin(ctx context.Context, m *meta.Meta[model.Analytic]) error
	GetByIdAdmin(ctx context.Context, id int) (*model.Analytic, error)
}

type AnalyticService struct {
	repositories repository.Repositories
}

func NewAnalyticService(repositories repository.Repositories) AnalyticService {
	return AnalyticService{repositories: repositories}
}

func (a AnalyticService) ListAdmin(ctx context.Context, m *meta.Meta[model.Analytic]) error {
	return a.repositories.AnalyticRepository.
		WithTable(model.Analytic{}.TableName()).
		Omit("request_body", "response_body").Paginate(ctx, m, ``)
}

func (a AnalyticService) GetByIdAdmin(ctx context.Context, id int) (*model.Analytic, error) {
	data, err := a.repositories.AnalyticRepository.
		WithTable(model.Analytic{}.TableName()).
		Find(ctx, "id = ?", id)
	if err != nil {
		return nil, err
	}

	return data, nil
}
