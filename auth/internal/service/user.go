package service

import (
	"context"
	"strings"

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
	ErrUserCannotBeDeleted      = apperr.New("user_cannot_be_deleted", apperr.WithTextTranslate(i18n.ErrUserCannotBeDeleted), apperr.WithCode(code.PermissionDenied))
)

type IUserService interface {
	Trx(db *gorm.DB) IUserService
	List(ctx context.Context, m *model2.Meta[model.User]) error
	GetById(ctx context.Context, id int) (*model.User, error)
	Create(ctx context.Context, input UserCreateInput) (*model.User, error)
	Update(ctx context.Context, id int, input UserUpdateInput) error
	Delete(ctx context.Context, id int) error
}

type UserService struct {
	userRepository     repository2.IUserRepository
	roleRepository     repository2.IRoleRepository
	userRoleRepository repository2.IUserRoleRepository
	userCache          cache2.IUserCache
	tokenCache         cache2.ITokenCache
	hasher             secret.Hasher
	db                 *gorm.DB
}

func NewUserService(
	userRepository repository2.IUserRepository,
	roleRepository repository2.IRoleRepository,
	userRoleRepository repository2.IUserRoleRepository,
	userCache cache2.IUserCache,
	tokenCache cache2.ITokenCache,
	hasher secret.Hasher,
) UserService {
	return UserService{
		userRepository:     userRepository,
		roleRepository:     roleRepository,
		userRoleRepository: userRoleRepository,
		userCache:          userCache,
		tokenCache:         tokenCache,
		hasher:             hasher,
	}
}

func (s UserService) Trx(db *gorm.DB) IUserService {
	s.userRepository = s.userRepository.Trx(db)
	s.userRoleRepository = s.userRoleRepository.Trx(db)
	s.db = db
	return s
}

func (s UserService) List(ctx context.Context, m *model2.Meta[model.User]) error {
	return s.userRepository.With("roles").Paginate(ctx, m, ``)
}

func (s UserService) GetById(ctx context.Context, id int) (*model.User, error) {
	user, err := s.userRepository.With("roles").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

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

func (s UserService) Create(ctx context.Context, input UserCreateInput) (*model.User, error) {
	password := s.hasher.HashString(input.Password)

	user, err := s.userRepository.Create(ctx, &model.User{
		Login:      strings.ToLower(input.Login),
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

	user, err = s.userRepository.With("roles").Find(ctx, `id = ?`, user.ID)
	if err != nil {
		return nil, err
	}

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

func (s UserService) Update(ctx context.Context, id int, input UserUpdateInput) error {
	user, err := s.userRepository.With("roles").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	var selects []interface{}
	if input.Login != nil && *input.Login != "" {
		user.Login = strings.ToLower(*input.Login)
		selects = append(selects, "login")
	}
	if input.FirstName != nil && *input.FirstName != "" {
		user.FirstName = *input.FirstName
		selects = append(selects, "first_name")
	}
	if input.Password != nil && *input.Password != "" {
		password := s.hasher.HashString(*input.Password)
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

	if len(selects) > 0 || input.Roles != nil {
		if input.Roles != nil || (input.Password != nil && *input.Password != "") {
			if err := s.tokenCache.DeleteAllUserTokens(ctx, user.ID); err != nil {
				return apperr.ErrInternal.WithError(err)
			}
		}
		if err := s.userCache.DelUsersInfo(user.ID).ChainExecDel(ctx); err != nil {
			return apperr.ErrInternal.WithError(err)
		}
	}

	return nil
}

func (s UserService) Delete(ctx context.Context, id int) error {
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

func (s UserService) createRoles(ctx context.Context, user *model.User, rls []int) error {
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