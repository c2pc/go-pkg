package model

type RoleName string

const (
	SuperAdmin RoleName = "SuperAdmin"
)

func (r RoleName) String() string {
	return string(r)
}

type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	RolePermissions []RolePermission `json:"role_permissions"`
}

func (m Role) TableName() string {
	return "auth_roles"
}

type RolePermission struct {
	RoleID       int  `json:"role_id"`
	PermissionID int  `json:"permission_id"`
	Read         bool `json:"read"`
	Write        bool `json:"write"`
	Exec         bool `json:"exec"`

	Permission *Permission `json:"auth_permission"`
}

func (m RolePermission) TableName() string {
	return "auth_role_permissions"
}

func IsRole(role string, r ...RoleName) bool {
	for _, r := range r {
		if role == string(r) {
			return true
		}
	}
	return false
}
