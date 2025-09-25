package seeders

import (
	"context"

	model2 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	repository2 "github.com/c2pc/go-pkg/v2/auth/internal/repository"
)

func RoleSeeder(ctx context.Context, roleRepository repository2.IRoleRepository, rolePermissionRepository repository2.IRolePermissionRepository, permissions []model2.Permission) (*model2.Role, error) {
	superAdmin, err := roleRepository.FirstOrCreate(ctx, &model2.Role{
		Name: model2.SuperAdmin,
	}, "id", `name = ?`, model2.SuperAdmin)
	if err != nil {
		return nil, err
	}

	broker, err := roleRepository.FirstOrCreate(ctx, &model2.Role{
		Name:        model2.Broker,
		LogDisabled: true,
	}, "id", `name = ?`, model2.Broker)
	if err != nil {
		return nil, err
	}

	if len(permissions) > 0 {
		permissionsIDs := make([]int, len(permissions))
		for i, permission := range permissions {
			permissionsIDs[i] = permission.ID
		}

		err = rolePermissionRepository.Delete(ctx, `role_id IN ? AND permission_id NOT IN (?)`, []int{superAdmin.ID, broker.ID}, permissionsIDs)
		if err != nil {
			return nil, err
		}

		for _, permission := range permissions {
			_, err = rolePermissionRepository.CreateOrUpdate(ctx, &model2.RolePermission{
				RoleID:       superAdmin.ID,
				PermissionID: permission.ID,
				Read:         true,
				Write:        true,
				Exec:         true,
			}, []interface{}{"role_id", "permission_id"}, []interface{}{"read", "write", "exec"}, nil)
			if err != nil {
				return nil, err
			}
			_, err = rolePermissionRepository.CreateOrUpdate(ctx, &model2.RolePermission{
				RoleID:       broker.ID,
				PermissionID: permission.ID,
				Read:         true,
				Write:        true,
				Exec:         true,
			}, []interface{}{"role_id", "permission_id"}, []interface{}{"read", "write", "exec"}, nil)
			if err != nil {
				return nil, err
			}
		}
	}

	return superAdmin, nil
}
