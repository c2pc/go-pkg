package cachekey

import (
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
)

const (
	UserInfo = "USER_INFO:"
)

func GetUserInfoKey(userID int64) string {
	return ServiceName + UserInfo + stringutil.Int64ToString(userID)
}
