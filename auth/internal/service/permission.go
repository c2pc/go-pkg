package service

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"gorm.io/gorm"
)

type IPermissionService interface {
	Trx(db *gorm.DB) IPermissionService
	List(ctx context.Context) ([]model.Permission, error)
}

type PermissionService struct {
	repositories repository.Repositories
	cache        *fx.CacheHolder
}

func NewPermissionService(
	repositories repository.Repositories,
	cache *fx.CacheHolder,
) PermissionService {
	return PermissionService{
		repositories: repositories,
		cache:        cache,
	}
}

func (s PermissionService) Trx(db *gorm.DB) IPermissionService {
	s.repositories.PermissionRepository = s.repositories.PermissionRepository.Trx(db)
	return s
}

func (s PermissionService) List(ctx context.Context) ([]model.Permission, error) {
	return s.cache.Get().PermissionCache.GetPermissionList(ctx, func(ctx context.Context) ([]model.Permission, error) {
		return s.repositories.PermissionRepository.GetList(ctx)
	})
}
