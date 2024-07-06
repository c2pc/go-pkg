package service

import (
	"context"
	"github.com/c2pc/go-pkg/v2/auth/cache"
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/repository"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"gorm.io/gorm"
)

type IPermissionService interface {
	Trx(db *gorm.DB) IPermissionService
	List(ctx context.Context) ([]model.Permission, error)
}

type PermissionService struct {
	permissionRepository repository.IPermissionRepository
	permissionCache      cache.IPermissionCache
}

func NewPermissionService(
	permissionRepository repository.IPermissionRepository,
	permissionCache cache.IPermissionCache,
) PermissionService {
	return PermissionService{
		permissionRepository: permissionRepository,
		permissionCache:      permissionCache,
	}
}

func (s PermissionService) Trx(db *gorm.DB) IPermissionService {
	s.permissionRepository = s.permissionRepository.Trx(db)
	return s
}

func (s PermissionService) List(ctx context.Context) ([]model.Permission, error) {
	return s.permissionCache.GetPermissionList(ctx, func(ctx context.Context) ([]model.Permission, error) {
		return s.permissionRepository.List(ctx, &model2.Filter{}, ``)
	})
}
