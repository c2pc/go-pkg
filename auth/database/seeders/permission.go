package seeders

import (
	"context"
	model2 "github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/repository"
	"github.com/c2pc/go-pkg/v2/utils/model"
)

func PermissionSeeder(ctx context.Context, permissionRepository repository.IPermissionRepository, permissions []string) ([]model2.Permission, error) {
	permissionsMap := make(map[string]struct{})
	for _, permission := range permissions {
		permissionsMap[permission] = struct{}{}
	}

	perms, err := permissionRepository.List(ctx, &model.Filter{}, ``)
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

	for permission, _ := range permissionsMap {
		if _, ok := permsMap[permission]; !ok {
			_, err := permissionRepository.Create(ctx, &model2.Permission{
				Name: permission,
			}, `id`)
			if err != nil {
				return nil, err
			}
		}
	}

	return permissionRepository.List(ctx, &model.Filter{}, ``)
}
