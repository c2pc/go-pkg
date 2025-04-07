package service

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/profile"

	cache2 "github.com/c2pc/go-pkg/v2/auth/internal/cache"
	"github.com/c2pc/go-pkg/v2/auth/internal/i18n"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	repository2 "github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound             = apperr.New("user_not_found", apperr.WithTextTranslate(i18n.ErrUserNotFound), apperr.WithCode(code.NotFound))
	ErrUserExists               = apperr.New("user_exists_error", apperr.WithTextTranslate(i18n.ErrUserExists), apperr.WithCode(code.InvalidArgument))
	ErrUserRolesCannotBeChanged = apperr.New("user_roles_cannot_be_changed", apperr.WithTextTranslate(i18n.ErrUserRolesCannotBeChanged), apperr.WithCode(code.PermissionDenied))
	ErrUserCannotBeBlocked      = apperr.New("user_cannot_be_blocked", apperr.WithTextTranslate(i18n.ErrUserCannotBeBlocked), apperr.WithCode(code.PermissionDenied))
	ErrUserCannotBeDeleted      = apperr.New("user_cannot_be_deleted", apperr.WithTextTranslate(i18n.ErrUserCannotBeDeleted), apperr.WithCode(code.PermissionDenied))
)

type IUserService[Model profile.IModel, CreateInput, UpdateInput, UpdateProfileInput any] interface {
	Trx(db *gorm.DB) IUserService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	List(ctx context.Context, m *model2.Meta[model.User]) error
	GetById(ctx context.Context, id int) (*model.User, error)
	Create(ctx context.Context, input UserCreateInput, profileInput *CreateInput) (*model.User, error)
	Update(ctx context.Context, id int, input UserUpdateInput, profileInput *UpdateInput) error
	Delete(ctx context.Context, id int) error
}

type UserService[Model profile.IModel, CreateInput, UpdateInput, UpdateProfileInput any] struct {
	profileService     profile.IProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]
	userRepository     repository2.IUserRepository
	roleRepository     repository2.IRoleRepository
	userRoleRepository repository2.IUserRoleRepository
	userCache          cache2.IUserCache
	tokenCache         cache2.ITokenCache
	hasher             secret.Hasher
	db                 *gorm.DB
}

func NewUserService[Model profile.IModel, CreateInput, UpdateInput, UpdateProfileInput any](
	profileService profile.IProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput],
	userRepository repository2.IUserRepository,
	roleRepository repository2.IRoleRepository,
	userRoleRepository repository2.IUserRoleRepository,
	userCache cache2.IUserCache,
	tokenCache cache2.ITokenCache,
	hasher secret.Hasher,
) UserService[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	return UserService[Model, CreateInput, UpdateInput, UpdateProfileInput]{
		profileService:     profileService,
		userRepository:     userRepository,
		roleRepository:     roleRepository,
		userRoleRepository: userRoleRepository,
		userCache:          userCache,
		tokenCache:         tokenCache,
		hasher:             hasher,
	}
}

func (s UserService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Trx(db *gorm.DB) IUserService[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	s.userRepository = s.userRepository.Trx(db)
	s.userRoleRepository = s.userRoleRepository.Trx(db)

	if s.profileService != nil {
		s.profileService = s.profileService.Trx(db)
	}

	return s
}

func (s UserService[Model, CreateInput, UpdateInput, UpdateProfileInput]) List(ctx context.Context, m *model2.Meta[model.User]) error {
	if err := s.userRepository.With("roles").Paginate(ctx, m, ``); err != nil {
		return err
	}

	if s.profileService != nil && len(m.Rows) > 0 {
		ids := make([]int, len(m.Rows))
		for i, user := range m.Rows {
			ids[i] = user.ID
		}

		profiles, err := s.profileService.GetByIds(ctx, ids...)
		if err != nil {
			return err
		}

		profilesMap := make(map[int]Model)

		for _, prof := range profiles {
			profilesMap[prof.GetUserId()] = prof
		}

		for i, user := range m.Rows {
			if prof, ok := profilesMap[user.ID]; ok {
				m.Rows[i].Profile = &prof
			}
		}
	}

	return nil
}

func (s UserService[Model, CreateInput, UpdateInput, UpdateProfileInput]) GetById(ctx context.Context, id int) (*model.User, error) {
	user, err := s.userRepository.With("roles").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	var prof *Model
	if s.profileService != nil {
		prof, err = s.profileService.GetById(ctx, user.ID)
		if err != nil {
			return nil, err
		}
	}

	user.Profile = prof

	return user, nil
}

type UserCreateInput struct {
	Login      string
	FirstName  string
	SecondName *string
	LastName   *string
	Password   string
	Email      *string
	Phone      *string
	Roles      []int
	Blocked    bool
}

func (s UserService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Create(ctx context.Context, input UserCreateInput, profileInput *CreateInput) (*model.User, error) {
	password, err := s.hasher.HashString(input.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepository.Create(ctx, &model.User{
		Login:      input.Login,
		FirstName:  input.FirstName,
		SecondName: input.SecondName,
		LastName:   input.LastName,
		Password:   password,
		Email:      input.Email,
		Phone:      input.Phone,
		Blocked:    input.Blocked,
	}, "id")
	if err != nil {
		if apperr.Is(err, apperr.ErrDBDuplicated) {
			return nil, ErrUserExists
		}
		return nil, err
	}

	if err := s.createRoles(ctx, user, input.Roles); err != nil {
		return nil, err
	}

	var prof *Model
	if s.profileService != nil && profileInput != nil {
		prof, err = s.profileService.Create(ctx, user.ID, *profileInput)
		if err != nil {
			return nil, err
		}
	}

	user, err = s.userRepository.With("roles").Find(ctx, `id = ?`, user.ID)
	if err != nil {
		return nil, err
	}

	user.Profile = prof

	return user, nil
}

type UserUpdateInput struct {
	Login      *string
	FirstName  *string
	SecondName *string
	LastName   *string
	Password   *string
	Email      *string
	Phone      *string
	Roles      []int
	Blocked    *bool
}

func (s UserService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Update(ctx context.Context, id int, input UserUpdateInput, profileInput *UpdateInput) error {
	user, err := s.userRepository.With("roles").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	var selects []interface{}
	if input.Login != nil && *input.Login != "" {
		user.Login = *input.Login
		selects = append(selects, "login")
	}
	if input.FirstName != nil && *input.FirstName != "" {
		user.FirstName = *input.FirstName
		selects = append(selects, "first_name")
	}
	if input.Password != nil && *input.Password != "" {
		password, err := s.hasher.HashString(*input.Password)
		if err != nil {
			return err
		}
		user.Password = password
		selects = append(selects, "password")
	}
	if input.SecondName != nil {
		if *input.SecondName == "" {
			user.SecondName = nil
		} else {
			user.SecondName = input.SecondName
		}
		selects = append(selects, "second_name")
	}
	if input.LastName != nil {
		if *input.LastName == "" {
			user.LastName = nil
		} else {
			user.LastName = input.LastName
		}
		selects = append(selects, "last_name")
	}
	if input.Email != nil {
		if *input.Email == "" {
			user.Email = nil
		} else {
			user.Email = input.Email
		}
		selects = append(selects, "email")
	}
	if input.Phone != nil {
		if *input.Phone == "" {
			user.Phone = nil
		} else {
			user.Phone = input.Phone
		}
		selects = append(selects, "phone")
	}
	if input.Blocked != nil {
		if *input.Blocked {
			var superAdminRole *model.Role
			for _, role := range user.Roles {
				if role.Name == model.SuperAdmin {
					superAdminRole = &role
					break
				}
			}

			if superAdminRole != nil {
				userIDs, err := s.userRoleRepository.GetUsersByRole(ctx, superAdminRole.ID)
				if err != nil {
					return err
				}

				if len(userIDs) <= 1 {
					return ErrUserCannotBeBlocked
				}
			}
		}

		user.Blocked = *input.Blocked
		selects = append(selects, "blocked")
	}

	if len(selects) > 0 {
		if err = s.userRepository.Update(ctx, user, selects, `id = ?`, user.ID); err != nil {
			if apperr.Is(err, apperr.ErrDBDuplicated) {
				return ErrUserExists
			}
			return err
		}
	}

	if input.Roles != nil {
		var superAdminRole *model.Role
		for _, role := range user.Roles {
			if role.Name == model.SuperAdmin {
				superAdminRole = &role
				break
			}
		}

		if superAdminRole != nil {
			userIDs, err := s.userRoleRepository.GetUsersByRole(ctx, superAdminRole.ID)
			if err != nil {
				return err
			}

			if len(userIDs) <= 1 {
				return ErrUserRolesCannotBeChanged
			}
		}

		if err = s.userRoleRepository.Delete(ctx, `user_id = ?`, user.ID); err != nil {
			return err
		}

		if err := s.createRoles(ctx, user, input.Roles); err != nil {
			return err
		}
	}

	if s.profileService != nil && profileInput != nil {
		err = s.profileService.Trx(s.db).Update(ctx, user.ID, *profileInput)
		if err != nil {
			return err
		}
	}

	if len(selects) > 0 || input.Roles != nil || profileInput != nil {
		if input.Roles != nil || (input.Password != nil && *input.Password != "") {
			if err := s.tokenCache.DeleteAllUserTokens(ctx, user.ID); err != nil {
				return apperr.ErrInternal.WithError(err)
			}
		}
		if err := s.userCache.DelUsersInfo(user.ID).ChainExecDel(ctx); err != nil {
			return apperr.ErrInternal.WithError(err)
		}
	}

	if input.Blocked != nil {
		if *input.Blocked {
			if err := s.userCache.DelUsersInfo(user.ID).ChainExecDel(ctx); err != nil {
				return apperr.ErrInternal.WithError(err)
			}

			if err := s.tokenCache.DeleteAllUserTokens(ctx, user.ID); err != nil {
				return apperr.ErrInternal.WithError(err)
			}
		}
	}

	return nil
}

func (s UserService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Delete(ctx context.Context, id int) error {
	user, err := s.userRepository.With("roles").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	superAdminRole := func() *model.Role {
		for _, role := range user.Roles {
			if role.Name == model.SuperAdmin {
				return &role
			}
		}
		return nil
	}()

	if superAdminRole != nil {
		userIDs, err := s.userRoleRepository.GetUsersByRole(ctx, superAdminRole.ID)
		if err != nil {
			return err
		}

		if len(userIDs) <= 1 {
			return ErrUserCannotBeDeleted
		}
	}

	if s.profileService != nil {
		err = s.profileService.Trx(s.db).Delete(ctx, user.ID)
		if err != nil {
			return err
		}
	}

	if err := s.userRepository.Delete(ctx, `id = ?`, user.ID); err != nil {
		return err
	}

	if err := s.userCache.DelUsersInfo(user.ID).ChainExecDel(ctx); err != nil {
		return apperr.ErrInternal.WithError(err)
	}

	if err := s.tokenCache.DeleteAllUserTokens(ctx, user.ID); err != nil {
		return apperr.ErrInternal.WithError(err)
	}

	return nil
}

func (s UserService[Model, CreateInput, UpdateInput, UpdateProfileInput]) createRoles(ctx context.Context, user *model.User, rls []int) error {
	if len(rls) > 0 {
		uniqueRoles := stringutil.RemoveDuplicate(rls)

		roles, err := s.roleRepository.List(ctx, &model2.Filter{}, `id IN (?)`, uniqueRoles)
		if err != nil {
			return err
		}

		var rolesToCreate []model.UserRole
		for _, role := range roles {
			rolesToCreate = append(rolesToCreate, model.UserRole{
				UserID: user.ID,
				RoleID: role.ID,
			})
		}

		if len(rolesToCreate) > 0 {
			if _, err := s.userRoleRepository.Create2(ctx, &rolesToCreate, ""); err != nil {
				return err
			}
		}
	}

	return nil
}
