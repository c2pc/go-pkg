package service

import (
	"context"
	"fmt"

	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/i18n"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/secret"
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
	"github.com/c2pc/go-pkg/v2/utils/syslog"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound             = apperr.New("user_not_found", apperr.WithTextTranslate(i18n.ErrUserNotFound), apperr.WithCode(code.NotFound))
	ErrUserExists               = apperr.New("user_exists_error", apperr.WithTextTranslate(i18n.ErrUserExists), apperr.WithCode(code.AlreadyExists))
	ErrUserRolesCannotBeChanged = apperr.New("user_roles_cannot_be_changed", apperr.WithTextTranslate(i18n.ErrUserRolesCannotBeChanged), apperr.WithCode(code.PermissionDenied))
	ErrUserCannotBeBlocked      = apperr.New("user_cannot_be_blocked", apperr.WithTextTranslate(i18n.ErrUserCannotBeBlocked), apperr.WithCode(code.PermissionDenied))
	ErrUserCannotBeDeleted      = apperr.New("user_cannot_be_deleted", apperr.WithTextTranslate(i18n.ErrUserCannotBeDeleted), apperr.WithCode(code.PermissionDenied))
	ErrSelfCannotBeDeleted      = apperr.New("self_cannot_be_deleted", apperr.WithTextTranslate(i18n.ErrSelfCannotBeDeleted), apperr.WithCode(code.PermissionDenied))

	ErrUserCannotCreateDomain = apperr.New("cannot_create_domain_admin", apperr.WithTextTranslate(i18n.ErrUserCannotCreateDomain), apperr.WithCode(code.PermissionDenied))
	ErrLocalCannotBeDomain    = apperr.New("local_cannot_be_domain", apperr.WithTextTranslate(i18n.ErrLocalCannotBeDomain), apperr.WithCode(code.PermissionDenied))
	ErrDomainCannotBeLocal    = apperr.New("domain_cannot_be_local", apperr.WithTextTranslate(i18n.ErrDomainCannotBeLocal), apperr.WithCode(code.PermissionDenied))
	ErrDomainLoginChange      = apperr.New("domain_login_change_forbidden", apperr.WithTextTranslate(i18n.ErrDomainLoginChange), apperr.WithCode(code.PermissionDenied))
	ErrDomainPasswordChange   = apperr.New("domain_password_change_forbidden", apperr.WithTextTranslate(i18n.ErrDomainPasswordChange), apperr.WithCode(code.PermissionDenied))
)

type IUserService interface {
	Trx(db *gorm.DB) IUserService
	List(ctx context.Context, m *meta.Meta[model.User]) error
	GetById(ctx context.Context, id int64) (*model.User, error)
	Create(ctx context.Context, input UserCreateInput, profileInput any) (*model.User, error)
	Update(ctx context.Context, id int64, input UserUpdateInput, profileInput any) (string, error)
	Delete(ctx context.Context, id int64) (string, error)
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

func (s UserService) List(ctx context.Context, m *meta.Meta[model.User]) error {
	if err := s.repositories.UserRepository.With("roles").Paginate(ctx, m, ``); err != nil {
		return err
	}

	if s.profileService != nil && len(m.Rows) > 0 {
		ids := make([]int64, len(m.Rows))
		for i, user := range m.Rows {
			ids[i] = user.ID
		}

		profiles, err := s.profileService.GetByIds(ctx, ids...)
		if err != nil {
			return err
		}

		profilesMap := make(map[int64]profile.IModel)

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

func (s UserService) GetById(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.repositories.UserRepository.With("roles").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	var prof profile.IModel
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

func (s UserService) Create(ctx context.Context, input UserCreateInput, profileInput any) (user *model.User, err error) {
	defer func() {
		success := err == nil
		msg := fmt.Sprintf("Создание учетной записи: %s", input.Login)
		if err != nil {
			msg = fmt.Sprintf("%s: %s", msg, err.Error())
		}

		syslog.Write(ctx, syslog.Record{EventID: "create-admin", EventName: "Создание учетной записи", Severity: syslog.SeverityLow, Success: success}, msg)

		if err == nil {
			if input.Blocked {
				syslog.Write(ctx, syslog.Record{EventID: "block-admin", EventName: "Блокировка учетной записи", Severity: syslog.SeverityLow, Success: success}, "Блокировка учетной записи: %s", input.Login)
			}
		}
	}()

	user, err = s.repositories.UserRepository.Create(ctx, &model.User{
		Login:      input.Login,
		FirstName:  input.FirstName,
		SecondName: input.SecondName,
		LastName:   input.LastName,
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

	var password *string
	if !input.IsDomain && input.Password != nil {
		pwd, err := secret.HasherSecret.HashString(model.GeneratePassword(*input.Password, user.ID))
		if err != nil {
			return nil, err
		}
		password = &pwd

		err = s.repositories.UserRepository.Update(ctx, &model.User{Password: password}, []interface{}{"password"}, `id = ?`, user.ID)
		if err != nil {
			return nil, err
		}
	}

	if err := s.createRoles(ctx, user, input.Roles); err != nil {
		return nil, err
	}

	var prof profile.IModel
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
	Login      *string
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

func (s UserService) Update(ctx context.Context, id int64, input UserUpdateInput, profileInput any) (userLogin string, err error) {
	var actions []syslog.Record
	var msgs []string
	defer func() {
		success := err == nil
		msg := fmt.Sprintf("Редактирование учетной записи: %s", userLogin)
		if err != nil {
			msg = fmt.Sprintf("%s: %s", msg, err.Error())
		}

		syslog.Write(ctx, syslog.Record{EventID: "update-admin", EventName: "Редактирование учетной записи", Severity: syslog.SeverityLow, Success: success}, msg)

		if err == nil {
			for i, act := range actions {
				act.Success = success
				syslog.Write(ctx, act, msgs[i])
			}
		}
	}()

	user, err := s.repositories.UserRepository.With("roles").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	userLogin = user.Login

	isDomain := user.IsDomain

	if input.IsDomain != nil {
		// Попытка сделать администратора доменным
		if *input.IsDomain && !user.IsDomain {
			isDomain = true
		}

		// Попытка сделать администратора недоменным
		if user.IsDomain && !*input.IsDomain {
			return userLogin, ErrDomainCannotBeLocal
		}
	}

	var selects []interface{}
	if input.IsDomain != nil && user.IsDomain != isDomain {
		user.IsDomain = isDomain
		selects = append(selects, "is_domain")
	}
	if input.Login != nil {
		if *input.Login != user.Login {
			if isDomain {
				return userLogin, ErrDomainLoginChange
			}

			user.Login = *input.Login
			selects = append(selects, "login")

			defer func() {
				actions = append(actions, syslog.Record{EventID: "update-admin-login", EventName: "Смена логина учетной записи", Severity: syslog.SeverityLow})
				msgs = append(msgs, fmt.Sprintf("Смена логина учетной записи: %s на %s", userLogin, *input.Login))
			}()
		}
	}
	if input.FirstName != nil && *input.FirstName != "" {
		user.FirstName = *input.FirstName
		selects = append(selects, "first_name")
	}
	if input.Password != nil {
		if isDomain {
			return userLogin, ErrDomainPasswordChange
		}

		if *input.Password == "" {
			user.Password = nil
		} else {
			password, err := secret.HasherSecret.HashString(model.GeneratePassword(*input.Password, user.ID))
			if err != nil {
				return userLogin, err
			}
			user.Password = &password
		}

		selects = append(selects, "password")

		actions = append(actions, syslog.Record{EventID: "update-admin-password", EventName: "Смена пароля учетной записи", Severity: syslog.SeverityLow})
		msgs = append(msgs, "")
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
			for _, r := range user.Roles {
				if model.IsRole(r.Name, model.SuperAdmin) {
					superAdminRole = &r
					break
				}
			}

			if superAdminRole != nil {
				userIDs, err := s.repositories.UserRoleRepository.GetUsersByRole(ctx, superAdminRole.ID)
				if err != nil {
					return userLogin, err
				}

				if len(userIDs) <= 1 {
					return userLogin, ErrUserCannotBeBlocked
				}
			}
		}

		if user.Blocked != *input.Blocked {
			if *input.Blocked {
				actions = append(actions, syslog.Record{EventID: "block-admin", EventName: "Блокировка учетной записи", Severity: syslog.SeverityLow})
				msgs = append(msgs, fmt.Sprintf("Блокировка учетной записи: %s", userLogin))
			} else {
				actions = append(actions, syslog.Record{EventID: "unlock-admin", EventName: "Разблокировка учетной записи", Severity: syslog.SeverityLow})
				msgs = append(msgs, fmt.Sprintf("Разблокировка учетной записи: %s", userLogin))
			}
		}

		user.Blocked = *input.Blocked
		selects = append(selects, "blocked")
	}

	if len(selects) > 0 {
		if err = s.repositories.UserRepository.Update(ctx, user, selects, `id = ?`, user.ID); err != nil {
			if apperr.Is(err, apperr.ErrDBDuplicated) {
				return userLogin, ErrUserExists
			}
			return userLogin, err
		}
	}

	if input.Roles != nil {
		var superAdminRole *model.Role
		for _, r := range user.Roles {
			if model.IsRole(r.Name, model.SuperAdmin) {
				superAdminRole = &r
				break
			}
		}

		if superAdminRole != nil {
			uniqueRoles := stringutil.RemoveDuplicate(input.Roles)

			roles, err := s.repositories.RoleRepository.List(ctx, &meta.Filter{}, `id IN (?)`, uniqueRoles)
			if err != nil {
				return userLogin, err
			}

			isSuperAdminRole := false
			for _, r := range roles {
				if model.IsRole(r.Name, model.SuperAdmin) {
					isSuperAdminRole = true
					break
				}
			}

			if !isSuperAdminRole {
				userIDs, err := s.repositories.UserRoleRepository.GetUsersByRole(ctx, superAdminRole.ID)
				if err != nil {
					return userLogin, err
				}

				if len(userIDs) <= 1 {
					return userLogin, ErrUserRolesCannotBeChanged
				}
			}
		}

		if err = s.repositories.UserRoleRepository.Delete(ctx, `user_id = ?`, user.ID); err != nil {
			return userLogin, err
		}

		if err := s.createRoles(ctx, user, input.Roles); err != nil {
			return userLogin, err
		}

		actions = append(actions, syslog.Record{EventID: "update-admin-roles", EventName: "Включение/Исключение пользователя в/из состава ролей", Severity: syslog.SeverityLow})
		msgs = append(msgs, fmt.Sprintf("Включение/Исключение пользователя в/из состава ролей: %s", userLogin))
	}

	if s.profileService != nil && profileInput != nil {
		err = s.profileService.Trx(s.db).Update(ctx, user.ID, profileInput)
		if err != nil {
			if !(apperr.Is(err, profile.ErrNotFound) ||
				apperr.Is(err, apperr.ErrDBRecordNotFound) ||
				apperr.Is(err, apperr.ErrNotFound)) {
				return userLogin, err
			}
		}
	}

	if len(selects) > 0 || input.Roles != nil || profileInput != nil {
		if err := s.cache.Get().UserCache.DelUsersInfo(user.ID).ChainExecDel(ctx); err != nil {
			return userLogin, apperr.ErrInternal.WithError(err)
		}
	}

	if input.Blocked != nil {
		if *input.Blocked {
			if err := s.cache.Get().UserCache.DelUsersInfo(user.ID).ChainExecDel(ctx); err != nil {
				return userLogin, apperr.ErrInternal.WithError(err)
			}

			if err := s.cache.Get().TokenCache.DeleteAllUserTokens(ctx, user.ID); err != nil {
				return userLogin, apperr.ErrInternal.WithError(err)
			}
		}
	}

	return userLogin, nil
}

func (s UserService) Delete(ctx context.Context, id int64) (login string, err error) {
	defer func() {
		success := err == nil
		msg := fmt.Sprintf("Удаление учетной записи %s", login)
		if err != nil {
			msg = fmt.Sprintf("%s: %s", msg, err.Error())
		}

		syslog.Write(ctx, syslog.Record{EventID: "delete-admin", EventName: "Удаление учетной записи", Severity: syslog.SeverityLow, Success: success}, msg)
	}()

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
			if model.IsRole(role.Name, model.SuperAdmin) {
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

		roles, err := s.repositories.RoleRepository.List(ctx, &meta.Filter{}, `id IN (?)`, uniqueRoles)
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
			if _, err := s.repositories.UserRoleRepository.Create2(ctx, &rolesToCreate, ""); err != nil {
				return err
			}
		}
	}

	return nil
}
