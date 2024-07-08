package service

import (
	"context"
	"github.com/c2pc/go-pkg/v2/auth/cache"
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"gorm.io/gorm"
	"slices"
)

var (
	ErrRoleNotFound = apperr.New("role_not_found",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Роль не найдена", translator.EN: "Role not found"}),
		apperr.WithCode(code.NotFound),
	)
	ErrRoleExists = apperr.New("role_exists_error",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Роль уже добавлена", translator.EN: "Role has already been added"}),
		apperr.WithCode(code.InvalidArgument),
	)
	ErrRoleCannotBeChanged = apperr.New("role_cannot_be_changed",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Роль нельзя редактировать", translator.EN: "Role cannot be changed"}),
		apperr.WithCode(code.PermissionDenied),
	)
	ErrRoleCannotBeDeleted = apperr.New("role_cannot_be_deleted",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Роль нельзя удалять", translator.EN: "Role cannot be deleted"}),
		apperr.WithCode(code.PermissionDenied),
	)
)

type IRoleService interface {
	Trx(db *gorm.DB) IRoleService
	List(ctx context.Context, m *model2.Meta[model.Role]) error
	GetById(ctx context.Context, id int) (*model.Role, error)
	Create(ctx context.Context, input RoleCreateInput) (*model.Role, error)
	Update(ctx context.Context, id int, input RoleUpdateInput) error
	Delete(ctx context.Context, id int) error
}

type RoleService struct {
	roleRepository           repository.IRoleRepository
	permissionRepository     repository.IPermissionRepository
	rolePermissionRepository repository.IRolePermissionRepository
	userRoleRepository       repository.IUserRoleRepository
	userCache                cache.IUserCache
}

func NewRoleService(
	roleRepository repository.IRoleRepository,
	permissionRepository repository.IPermissionRepository,
	rolePermissionRepository repository.IRolePermissionRepository,
	userRoleRepository repository.IUserRoleRepository,
	userCache cache.IUserCache,
) RoleService {
	return RoleService{
		roleRepository:           roleRepository,
		permissionRepository:     permissionRepository,
		rolePermissionRepository: rolePermissionRepository,
		userRoleRepository:       userRoleRepository,
		userCache:                userCache,
	}
}

func (s RoleService) Trx(db *gorm.DB) IRoleService {
	s.roleRepository = s.roleRepository.Trx(db)
	return s
}

func (s RoleService) List(ctx context.Context, m *model2.Meta[model.Role]) error {
	return s.roleRepository.With("role_permissions").Paginate(ctx, m, ``)
}

func (s RoleService) GetById(ctx context.Context, id int) (*model.Role, error) {
	role, err := s.roleRepository.With("role_permissions").Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	return role, nil
}

type RolePermissions struct {
	Write []int
	Read  []int
	Exec  []int
}

type RoleCreateInput struct {
	Name        string
	Permissions RolePermissions
}

func (s RoleService) Create(ctx context.Context, input RoleCreateInput) (*model.Role, error) {
	role, err := s.roleRepository.Create(ctx, &model.Role{
		Name: input.Name,
	}, "id")
	if err != nil {
		if apperr.Is(err, apperr.ErrDBDuplicated) {
			return nil, ErrRoleExists
		}
		return nil, err
	}

	rolePermissions, err := s.createPermissions(ctx, role, input.Permissions)
	if err != nil {
		return nil, err
	}

	role.RolePermissions = rolePermissions

	return role, nil
}

type RoleUpdateInput struct {
	Name        *string
	Permissions *RolePermissions
}

func (s RoleService) Update(ctx context.Context, id int, input RoleUpdateInput) error {
	role, err := s.roleRepository.Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	if role.Name == model.SuperAdmin {
		return ErrRoleCannotBeChanged
	}

	if input.Name != nil && *input.Name != "" {
		if *input.Name != role.Name {
			if err = s.roleRepository.Update(ctx, &model.Role{Name: *input.Name}, []interface{}{"name"}, `id = ?`, role.ID); err != nil {
				if apperr.Is(err, apperr.ErrDBDuplicated) {
					return ErrRoleExists
				}
				return err
			}
		}
	}

	if input.Permissions != nil {
		if err = s.rolePermissionRepository.Delete(ctx, `role_id = ?`, role.ID); err != nil {
			return err
		}

		_, err = s.createPermissions(ctx, role, *input.Permissions)
		if err != nil {
			return err
		}

		userIDs, err := s.userRoleRepository.GetUsersByRole(ctx, role.ID)
		if err != nil {
			return err
		}

		if len(userIDs) > 0 {
			if err := s.userCache.DelUsersInfo(userIDs...).ChainExecDel(ctx); err != nil {
				return apperr.ErrInternal.WithError(err)
			}
		}
	}

	return nil
}

func (s RoleService) Delete(ctx context.Context, id int) error {
	role, err := s.roleRepository.Find(ctx, `id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrRoleNotFound
		}
		return err
	}

	userIDs, err := s.userRoleRepository.GetUsersByRole(ctx, role.ID)
	if err != nil {
		return err
	}

	if role.Name == model.SuperAdmin {
		return ErrRoleCannotBeDeleted
	}

	if err := s.roleRepository.Delete(ctx, `id = ?`, role.ID); err != nil {
		return err
	}

	if len(userIDs) > 0 {
		if err := s.userCache.DelUsersInfo(userIDs...).ChainExecDel(ctx); err != nil {
			return apperr.ErrInternal.WithError(err)
		}
	}

	return nil
}

func (s RoleService) createPermissions(ctx context.Context, role *model.Role, perms RolePermissions) ([]model.RolePermission, error) {
	var rolePermissions []model.RolePermission

	if len(perms.Write) > 0 || len(perms.Read) > 0 || len(perms.Exec) > 0 {
		write := stringutil.RemoveDuplicate(perms.Write)
		read := stringutil.RemoveDuplicate(perms.Read)
		exec := stringutil.RemoveDuplicate(perms.Exec)

		uniquePerms := stringutil.RemoveDuplicate(slices.Concat(write, read, exec))

		permissions, err := s.permissionRepository.List(ctx, &model2.Filter{}, `id IN (?)`, uniquePerms)
		if err != nil {
			return nil, err
		}

		permissionsToCreate := make(map[int]model.RolePermission)
		for _, permission := range permissions {
			permissionsToCreate[permission.ID] = model.RolePermission{
				RoleID:       role.ID,
				PermissionID: permission.ID,
				Read:         false,
				Write:        false,
				Exec:         false,
			}
		}

		for _, w := range write {
			if v, ok := permissionsToCreate[w]; ok {
				v.Write = true
				permissionsToCreate[w] = v
			}
		}

		for _, w := range read {
			if v, ok := permissionsToCreate[w]; ok {
				v.Read = true
				permissionsToCreate[w] = v
			}
		}

		for _, w := range exec {
			if v, ok := permissionsToCreate[w]; ok {
				v.Exec = true
				permissionsToCreate[w] = v
			}
		}

		for _, v := range permissionsToCreate {
			rolePermissions = append(rolePermissions, v)
		}

		if len(rolePermissions) > 0 {
			if _, err := s.rolePermissionRepository.Create2(ctx, &rolePermissions, ""); err != nil {
				return nil, err
			}
		}
	}

	return rolePermissions, nil
}