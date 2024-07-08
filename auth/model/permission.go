package model

import (
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/translator"
)

var Permissions = []model2.Permission{
	{Method: "auth/roles", Desc: translator.Translate{translator.RU: "Роли", translator.EN: "Roles"}},
	{Method: "auth/users", Desc: translator.Translate{translator.RU: "Пользователи", translator.EN: "Users"}},
}

var permissions = make(map[string]translator.Translate)

func init() {
	for _, p := range Permissions {
		permissions[p.Method] = p.Desc
	}
}

func SetPermissions(perms []model2.Permission) {
	for _, p := range perms {
		permissions[p.Method] = p.Desc
	}
}

func GetPermissions() map[string]translator.Translate {
	return permissions
}

func GetPermissionsKeys() []string {
	keys := make([]string, 0, len(permissions))
	for k := range permissions {
		keys = append(keys, k)
	}
	return keys
}

func GetPermission(key string) translator.Translate {
	return permissions[key]
}

type Permission struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (m Permission) TableName() string {
	return "auth_permissions"
}
