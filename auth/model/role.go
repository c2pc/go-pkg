package model

const SuperAdmin = "super_admin"

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	RolePermissions []RolePermission `json:"role_permissions"`
}

func (m Role) TableName() string {
	return "roles"
}

type RolePermission struct {
	RoleID       int  `json:"role_id"`
	PermissionID int  `json:"permission_id"`
	Read         bool `json:"read"`
	Write        bool `json:"write"`
	Exec         bool `json:"exec"`

	Permission *Permission `json:"permission"`
}

func (m RolePermission) TableName() string {
	return "role_permissions"
}