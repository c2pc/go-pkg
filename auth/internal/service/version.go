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
	version             string
	migrationRepository repository.IMigrationRepository
}

func NewVersionService(version string, migrationRepository repository.IMigrationRepository) VersionService {
	return VersionService{
		version:             version,
		migrationRepository: migrationRepository,
	}
}

func (s VersionService) Get(ctx context.Context) *model.Version {
	version := &model.Version{
		App: s.version,
		DB:  "0.0.0",
	}
	m, _ := s.migrationRepository.Find(ctx, `version IS NOT NULL`)
	if m != nil {
		version.DB = "0.0." + m.Version
	}

	return version
}
