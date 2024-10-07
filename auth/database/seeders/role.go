package seeders

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/repository"
)

func RoleSeeder(ctx context.Context, roleRepository repository.IRoleRepository, rolePermissionRepository repository.IRolePermissionRepository, permissions []model.Permission) (*model.Role, error) {
	role, err := roleRepository.FirstOrCreate(ctx, &model.Role{
		Name: model.SuperAdmin,
	}, "id", `name = ?`, model.SuperAdmin)
	if err != nil {
		return nil, err
	}

	if len(permissions) > 0 {
		permissionsIDs := make([]int, len(permissions))
		for i, permission := range permissions {
			permissionsIDs[i] = permission.ID
		}

		err = rolePermissionRepository.Delete(ctx, `role_id = ? AND permission_id NOT IN (?)`, role.ID, permissionsIDs)
		if err != nil {
			return nil, err
		}

		for _, permission := range permissions {
			_, err = rolePermissionRepository.CreateOrUpdate(ctx, &model.RolePermission{
				RoleID:       role.ID,
				PermissionID: permission.ID,
				Read:         true,
				Write:        true,
				Exec:         true,
			}, []interface{}{"role_id", "permission_id"}, []interface{}{"read", "write", "exec"})
			if err != nil {
				return nil, err
			}
		}
	}

	return role, nil
}
