package repository

import (
	"gorm.io/gorm"
)

type Repositories struct {
	UserRepository           IUserRepository
	TokenRepository          ITokenRepository
	RoleRepository           IRoleRepository
	PermissionRepository     IPermissionRepository
	RolePermissionRepository IRolePermissionRepository
	UserRoleRepository       IUserRoleRepository
	SettingRepository        ISettingRepository
	FilterRepository         IFilterRepository
}

func NewRepositories(db *gorm.DB) Repositories {
	return Repositories{
		UserRepository:           NewUserRepository(db),
		TokenRepository:          NewTokenRepository(db),
		RoleRepository:           NewRoleRepository(db),
		PermissionRepository:     NewPermissionRepository(db),
		RolePermissionRepository: NewRolePermissionRepository(db),
		UserRoleRepository:       NewUserRoleRepository(db),
		SettingRepository:        NewSettingRepository(db),
		FilterRepository:         NewFilterRepository(db),
	}
}
