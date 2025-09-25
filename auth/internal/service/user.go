package service

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/i18n"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound             = apperr.New("user_not_found", apperr.WithTextTranslate(i18n.ErrUserNotFound), apperr.WithCode(code.NotFound))
	ErrUserExists               = apperr.New("user_exists_error", apperr.WithTextTranslate(i18n.ErrUserExists), apperr.WithCode(code.AlreadyExists))
	ErrUserRolesCannotBeChanged = apperr.New("user_roles_cannot_be_changed", apperr.WithTextTranslate(i18n.ErrUserRolesCannotBeChanged), apperr.WithCode(code.PermissionDenied))
	ErrUserCannotBeBlocked      = apperr.New("user_cannot_be_blocked", apperr.WithTextTranslate(i18n.ErrUserCannotBeBlocked), apperr.WithCode(code.PermissionDenied))
	ErrUserCannotBeDeleted      = apperr.New("user_cannot_be_deleted", apperr.WithTextTranslate(i18n.ErrUserCannotBeDeleted), apperr.WithCode(code.PermissionDenied))
	ErrSelfCannotBeDeleted      = apperr.New("self_cannot_be_deleted", apperr.WithTextTranslate(i18n.ErrSelfCannotBeDeleted), apperr.WithCode(code.PermissionDenied))
)

type IUserService interface {
	Trx(db *gorm.DB) IUserService
	List(ctx context.Context, m *model2.Meta[model.User]) error
	GetById(ctx context.Context, id int) (*model.User, error)
	Create(ctx context.Context, input UserCreateInput, profileInput any) (*model.User, error)
	Update(ctx context.Context, id int, input UserUpdateInput, profileInput any) (string, error)
	Delete(ctx context.Context, id int) (string, error)
}

type UserService struct {
	profileService profile.IProfileService
	repositories   repository.Repositories
	cache          *fx.CacheHolder
	db             *gorm.DB
}

func NewUserService(
	profileService profile.IProfileService,
	repositories repository.Repositories,
	cache *fx.CacheHolder,
) UserService {
	return UserService{
		profileService: profileService,
		repositories:   repositories,
		cache:          cache,
	}
}

func (s UserService) Trx(db *gorm.DB) IUserService {
	s.repositories.UserRepository = s.repositories.UserRepository.Trx(db)
	s.repositories.UserRoleRepository = s.repositories.UserRoleRepository.Trx(db)

	if s.profileService != nil {
		s.profileService = s.profileService.Trx(db)
	}

	return s
}

func (s UserService) List(ctx context.Context, m *model2.Meta[model.User]) error {
	if err := s.repositories.UserRepository.With("roles").Paginate(ctx, m, ``); err != nil {
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

		profilesMap := make(map[int]profile.IModel)

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

func (s UserService) GetById(ctx context.Context, id int) (*model.User, error) {
	user, err := s.repositories.UserRepository.With("roles").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	var prof *profile.IModel
	if s.profileService != nil {
		prof, err = s.profileService.GetById(ctx, user.ID)
		if err != nil {
			if !(apperr.Is(err, profile.ErrNotFound) ||
				apperr.Is(err, apperr.ErrDBRecordNotFound) ||
				apperr.Is(err, apperr.ErrNotFound)) {
				return nil, apperr.ErrUnauthenticated.WithError(err)
			}
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
	Password   *string
	Email      *string
	Phone      *string
	Roles      []int
	Blocked    bool
	IsDomain   bool
}

func (s UserService) Create(ctx context.Context, input UserCreateInput, profileInput any) (*model.User, error) {
	role, ok := mcontext.GetOpUserRole(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation user role is empty")
	}

	isBroker := role == model.Broker

	if !isBroker && input.IsDomain {
		return nil, apperr.ErrForbidden.WithErrorText("user is a domain")
	}

	var password *string
	if !input.IsDomain && input.Password != nil {
		pwd, err := secret.HasherSecret.HashString(*input.Password)
		if err != nil {
			return nil, err
		}
		password = &pwd
	}

	user, err := s.repositories.UserRepository.Create(ctx, &model.User{
		Login:      input.Login,
		FirstName:  input.FirstName,
		SecondName: input.SecondName,
		LastName:   input.LastName,
		Password:   password,
		Email:      input.Email,
		Phone:      input.Phone,
		Blocked:    input.Blocked,
		IsDomain:   input.IsDomain,
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

	var prof *profile.IModel
	if s.profileService != nil && profileInput != nil {
		prof, err = s.profileService.Create(ctx, user.ID, profileInput)
		if err != nil {
			if !(apperr.Is(err, profile.ErrNotFound) ||
				apperr.Is(err, apperr.ErrDBRecordNotFound) ||
				apperr.Is(err, apperr.ErrNotFound)) {
				return nil, err
			}
		}
	}

	user, err = s.repositories.UserRepository.With("roles").Find(ctx, `id = ?`, user.ID)
	if err != nil {
		return nil, err
	}

	user.Profile = prof

	return user, nil
}

type UserUpdateInput struct {
	FirstName  *string
	SecondName *string
	LastName   *string
	Password   *string
	Email      *string
	Phone      *string
	Roles      []int
	Blocked    *bool
	IsDomain   *bool
}

func (s UserService) Update(ctx context.Context, id int, input UserUpdateInput, profileInput any) (string, error) {
	user, err := s.repositories.UserRepository.With("roles").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	role, ok := mcontext.GetOpUserRole(ctx)
	if !ok {
		return user.Login, apperr.ErrUnauthenticated.WithErrorText("operation user role is empty")
	}

	isBroker := role == model.Broker
	isDomain := user.IsDomain
	if input.IsDomain != nil {
		isDomain = *input.IsDomain
	}

	var selects []interface{}
	if input.IsDomain != nil {
		if !isBroker {
			return user.Login, apperr.ErrForbidden.WithErrorText("user is a domain")
		}

		user.IsDomain = *input.IsDomain
		selects = append(selects, "is_domain")
	}
	if input.FirstName != nil && *input.FirstName != "" {
		user.FirstName = *input.FirstName
		selects = append(selects, "first_name")
	}
	if input.Password != nil {
		if isDomain {
			user.Password = nil
		} else if *input.Password == "" {
			user.Password = nil
		} else {
			password, err := secret.HasherSecret.HashString(*input.Password)
			if err != nil {
				return user.Login, err
			}
			user.Password = &password
		}

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
		if !isBroker && isDomain {
			return user.Login, apperr.ErrForbidden.WithErrorText("user is a domain")
		}

		if *input.Blocked {
			var superAdminRole *model.Role
			for _, role := range user.Roles {
				if role.Name == model.SuperAdmin {
					superAdminRole = &role
					break
				}
			}

			if superAdminRole != nil {
				userIDs, err := s.repositories.UserRoleRepository.GetUsersByRole(ctx, superAdminRole.ID)
				if err != nil {
					return user.Login, err
				}

				if len(userIDs) <= 1 {
					return user.Login, ErrUserCannotBeBlocked
				}
			}
		}

		user.Blocked = *input.Blocked
		selects = append(selects, "blocked")
	}

	if len(selects) > 0 {
		if err = s.repositories.UserRepository.Update(ctx, user, selects, `id = ?`, user.ID); err != nil {
			if apperr.Is(err, apperr.ErrDBDuplicated) {
				return user.Login, ErrUserExists
			}
			return user.Login, err
		}
	}

	if input.Roles != nil {
		if !isBroker && isDomain {
			return user.Login, apperr.ErrForbidden.WithErrorText("user is a domain")
		}

		var superAdminRole *model.Role
		for _, r := range user.Roles {
			if r.Name == model.SuperAdmin {
				superAdminRole = &r
				break
			}
		}

		if superAdminRole != nil {
			uniqueRoles := stringutil.RemoveDuplicate(input.Roles)

			roles, err := s.repositories.RoleRepository.List(ctx, &model2.Filter{}, `id IN (?)`, uniqueRoles)
			if err != nil {
				return user.Login, err
			}

			isSuperAdminRole := false
			for _, r := range roles {
				if r.Name == model.SuperAdmin {
					isSuperAdminRole = true
					break
				}
			}

			if !isSuperAdminRole {
				userIDs, err := s.repositories.UserRoleRepository.GetUsersByRole(ctx, superAdminRole.ID)
				if err != nil {
					return user.Login, err
				}

				if len(userIDs) <= 1 {
					return user.Login, ErrUserRolesCannotBeChanged
				}
			}
		}

		if err = s.repositories.UserRoleRepository.Delete(ctx, `user_id = ?`, user.ID); err != nil {
			return user.Login, err
		}

		if err := s.createRoles(ctx, user, input.Roles); err != nil {
			return user.Login, err
		}
	}

	if s.profileService != nil && profileInput != nil {
		err = s.profileService.Trx(s.db).Update(ctx, user.ID, profileInput)
		if err != nil {
			if !(apperr.Is(err, profile.ErrNotFound) ||
				apperr.Is(err, apperr.ErrDBRecordNotFound) ||
				apperr.Is(err, apperr.ErrNotFound)) {
				return user.Login, err
			}
		}
	}

	if len(selects) > 0 || input.Roles != nil || profileInput != nil {
		if err := s.cache.Get().UserCache.DelUsersInfo(user.ID).ChainExecDel(ctx); err != nil {
			return user.Login, apperr.ErrInternal.WithError(err)
		}
	}

	if input.Blocked != nil {
		if *input.Blocked {
			if err := s.cache.Get().UserCache.DelUsersInfo(user.ID).ChainExecDel(ctx); err != nil {
				return user.Login, apperr.ErrInternal.WithError(err)
			}

			if err := s.cache.Get().TokenCache.DeleteAllUserTokens(ctx, user.ID); err != nil {
				return user.Login, apperr.ErrInternal.WithError(err)
			}
		}
	}

	return user.Login, nil
}

func (s UserService) Delete(ctx context.Context, id int) (string, error) {
	user, err := s.repositories.UserRepository.With("roles").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}

	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return user.Login, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	if user.ID == userID {
		return user.Login, ErrSelfCannotBeDeleted
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
		userIDs, err := s.repositories.UserRoleRepository.GetUsersByRole(ctx, superAdminRole.ID)
		if err != nil {
			return user.Login, err
		}

		if len(userIDs) <= 1 {
			return user.Login, ErrUserCannotBeDeleted
		}
	}

	if s.profileService != nil {
		err = s.profileService.Trx(s.db).Delete(ctx, user.ID)
		if err != nil {
			if !(apperr.Is(err, profile.ErrNotFound) ||
				apperr.Is(err, apperr.ErrDBRecordNotFound) ||
				apperr.Is(err, apperr.ErrNotFound)) {
				return user.Login, err
			}
		}
	}

	if err := s.repositories.UserRepository.Delete(ctx, `id = ?`, user.ID); err != nil {
		return user.Login, err
	}

	if err := s.cache.Get().UserCache.DelUsersInfo(user.ID).ChainExecDel(ctx); err != nil {
		return user.Login, apperr.ErrInternal.WithError(err)
	}

	if err := s.cache.Get().TokenCache.DeleteAllUserTokens(ctx, user.ID); err != nil {
		return user.Login, apperr.ErrInternal.WithError(err)
	}

	return user.Login, nil
}

func (s UserService) createRoles(ctx context.Context, user *model.User, rls []int) error {
	if len(rls) > 0 {
		uniqueRoles := stringutil.RemoveDuplicate(rls)

		roles, err := s.repositories.RoleRepository.List(ctx, &model2.Filter{}, `id IN (?)`, uniqueRoles)
		if err != nil {
			return err
		}

		var rolesToCreate []model.UserRole
		for _, role := range roles {
			if role.Name == model.SuperAdmin || role.Name == model.Broker {
				rolesToCreate = []model.UserRole{
					{
						UserID: user.ID,
						RoleID: role.ID,
					},
				}
				break
			}

			rolesToCreate = append(rolesToCreate, model.UserRole{
				UserID: user.ID,
				RoleID: role.ID,
			})
		}

		if len(rolesToCreate) > 0 {
			if _, err := s.repositories.UserRoleRepository.Create2(ctx, &rolesToCreate, ""); err != nil {
				return err
			}
		}
	}

	return nil
}
