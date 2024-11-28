package seeders

import (
	"context"
	"strings"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	repository2 "github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/secret"
)

func UserSeeder(ctx context.Context, userRepository repository2.IUserRepository, userRoleRepository repository2.IUserRoleRepository, hasher secret.Hasher, roleID int) (*model.User, error) {
	role, err := userRoleRepository.Find(ctx, `role_id = ?`, roleID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			login := "admin"
			name := "Admin"
			password := "admin"
			pass := hasher.HashString(password)

			admin, err := userRepository.FirstOrCreate(ctx, &model.User{
				Login:     strings.ToLower(login),
				FirstName: name,
				Password:  pass,
			}, "id", `login = ?`, login)
			if err != nil {
				return nil, err
			}

			_, err = userRoleRepository.FirstOrCreate(ctx, &model.UserRole{
				UserID: admin.ID,
				RoleID: roleID,
			}, "", `user_id = ? AND role_id = ?`, admin.ID, roleID)
			if err != nil {
				return nil, err
			}

			return admin, nil
		}
		return nil, err
	}

	admin, err := userRepository.Find(ctx, "id", role.UserID)
	if err != nil {
		return nil, err
	}

	return admin, nil
}
