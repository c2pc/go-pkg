package seeders

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/meta"
)

func PermissionSeeder(ctx context.Context, permissionRepository repository.IPermissionRepository, permissions []string) ([]model.Permission, error) {
	permissionsMap := make(map[string]struct{})
	for _, permission := range permissions {
		permissionsMap[permission] = struct{}{}
	}

	perms, err := permissionRepository.List(ctx, &meta.Filter{}, ``)
	if err != nil {
		return nil, err
	}

	permsMap := make(map[string]struct{})
	for _, perm := range perms {
		permsMap[perm.Name] = struct{}{}
		if _, ok := permissionsMap[perm.Name]; !ok {
			err := permissionRepository.Delete(ctx, `id = ?`, perm.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	for permission := range permissionsMap {
		if _, ok := permsMap[permission]; !ok {
			_, err := permissionRepository.Create(ctx, &model.Permission{
				Name: permission,
			}, `id`)
			if err != nil {
				return nil, err
			}
		}
	}

	return permissionRepository.List(ctx, &meta.Filter{}, ``)
}
