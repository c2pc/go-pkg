package cachekey

const (
	PermissionList = "PERMISSION_LIST"
)

func GetPermissionListKey() string {
	return ServiceName + PermissionList
}
