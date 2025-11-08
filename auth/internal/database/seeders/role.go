package seeders

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
)

func RoleSeeder(ctx context.Context, roleRepository repository.IRoleRepository, rolePermissionRepository repository.IRolePermissionRepository, permissions []model.Permission) (*model.Role, error) {
	superAdmin, err := roleRepository.FirstOrCreate(ctx, &model.Role{
		Name: model.SuperAdmin.String(),
	}, "id", `name = ?`, model.SuperAdmin.String())
	if err != nil {
		return nil, err
	}

	if len(permissions) > 0 {
		var permsToCreate []model.RolePermission
		for _, permission := range permissions {
			permsToCreate = append(permsToCreate,
				model.RolePermission{
					RoleID:       superAdmin.ID,
					PermissionID: permission.ID,
					Read:         true,
					Write:        true,
					Exec:         true,
				})
		}

		err = rolePermissionRepository.Delete(ctx, `role_id IN ?`, []int{superAdmin.ID})
		if err != nil {
			return nil, err
		}

		if len(permsToCreate) > 0 {
			_, err = rolePermissionRepository.Create2(ctx, &permsToCreate)
			if err != nil {
				return nil, err
			}
		}
	}

	return superAdmin, nil
}
