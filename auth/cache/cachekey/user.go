package cachekey

import (
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
)

const (
	UserInfo = "USER_INFO:"
)

func GetUserInfoKey(userID int) string {
	return UserInfo + stringutil.IntToString(userID)
}
