package service

import (
	"context"
	"fmt"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/app_data"
)

var dbVersion string

type IVersionService interface {
	Get(ctx context.Context) *model.Version
}

type VersionService struct {
	version string

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
		AppName: app_data.AppName,
		App:     app_data.AppVersion,
		DB:      "0.0.0",
	}

	if dbVersion == "" {
		var serviceVersion, authVersion string
		m, _ := s.repositories.MigrationRepository.Find(ctx, `version IS NOT NULL`)
		if m != nil {
			serviceVersion = m.Version
		} else {
			serviceVersion = "0"
		}

		m2, _ := s.repositories.MigrationRepository.WithTable(model.Migration{}.TableNameAuth()).Find(ctx, `version IS NOT NULL`)
		if m2 != nil {
			authVersion = m2.Version
		} else {
			authVersion = "0"
		}

		dbVersion = fmt.Sprintf("0.%s.%s", authVersion, serviceVersion)
		version.DB = dbVersion
	} else {
		version.DB = dbVersion
	}

	return version
}
