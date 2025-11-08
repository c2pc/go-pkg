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
	MigrationRepository      IMigrationRepository
	ConfigRepository         IConfigRepository
	ConfigFileRepository     IConfigFileRepository
	AnalyticRepository       IAnalyticRepository
	UserBlockedRepository    IUserBlockedRepository
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
		MigrationRepository:      NewMigrationRepository(db),
		ConfigRepository:         NewConfigRepository(db),
		ConfigFileRepository:     NewConfigFileRepository(db),
		AnalyticRepository:       NewAnalyticRepository(db),
		UserBlockedRepository:    NewUserBlockedRepository(db),
	}
}
