package service

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
)

type IVersionService interface {
	Get(ctx context.Context) *model.Version
}

type VersionService struct {
	version      string
	repositories repository.Repositories
}

func NewVersionService(version string, repositories repository.Repositories) VersionService {
	return VersionService{
		version:      version,
		repositories: repositories,
	}
}

func (s VersionService) Get(ctx context.Context) *model.Version {
	version := &model.Version{
		App: s.version,
		DB:  "0.0.0",
	}
	m, _ := s.repositories.MigrationRepository.Find(ctx, `version IS NOT NULL`)
	if m != nil {
		version.DB = "0.0." + m.Version
	}

	return version
}
