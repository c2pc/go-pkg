package seeders

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	repository2 "github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/secret"
)

func UserSeeder(ctx context.Context, userRepository repository2.IUserRepository, userRoleRepository repository2.IUserRoleRepository, roleID int) (*model.User, error) {
	login := "admin"
	name := "Admin"
	password := "admin"

	role, err := userRoleRepository.Find(ctx, `role_id = ?`, roleID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			admin, err := userRepository.FirstOrCreate(ctx, &model.User{
				Login:     login,
				FirstName: name,
				IsDomain:  false,
			}, "id", `login = ?`, login)
			if err != nil {
				return nil, err
			}

			pass, err := secret.HasherSecret.HashString(model.GeneratePassword(password, admin.ID))
			if err != nil {
				return nil, err
			}

			err = userRepository.Update(ctx, &model.User{Password: &pass}, []interface{}{"password"}, `id = ?`, admin.ID)
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

	admin, err := userRepository.Find(ctx, "login = ?", login)
	if err == nil {
		if admin.Password != nil {
			if secret.HasherSecret.HashMatchesString(*admin.Password, password) {
				pass, err := secret.HasherSecret.HashString(model.GeneratePassword(password, admin.ID))
				if err != nil {
					return nil, err
				}
				err = userRepository.Update(ctx, &model.User{Password: &pass}, []interface{}{"password"}, `id = ?`, admin.ID)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	admin, err = userRepository.Find(ctx, "id", role.UserID)
	if err != nil {
		return nil, err
	}

	return admin, nil
}
