package seeders

import (
	"context"
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/secret"
)

func UserSeeder(ctx context.Context, userRepository repository.IUserRepository, userRoleRepository repository.IUserRoleRepository, hasher secret.Hasher, roleID int) error {
	_, err := userRoleRepository.Find(ctx, `role_id = ?`, roleID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			login := "admin"
			name := "Admin"
			password := "admin"
			pass := hasher.HashString(password)

			admin, err := userRepository.FirstOrCreate(ctx, &model.User{
				Login:     login,
				FirstName: name,
				Password:  pass,
			}, "id", `login = ?`, login)
			if err != nil {
				return err
			}

			_, err = userRoleRepository.FirstOrCreate(ctx, &model.UserRole{
				UserID: admin.ID,
				RoleID: roleID,
			}, "", `user_id = ? AND role_id = ?`, admin.ID, roleID)
			if err != nil {
				return err
			}

			return nil
		}
		return err
	}

	return nil
}
