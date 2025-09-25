package cachekey

import (
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
)

const (
	UserInfo = "USER_ANALYTICS:"
)

func GetUserInfoKey(userID int) string {
	return ServiceName + UserInfo + stringutil.IntToString(userID)
}
