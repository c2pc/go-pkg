package cachekey

import (
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
)

const (
	UidPidToken = "TOKEN_STATUS:"
)

func GetTokenKey(userID int, DeviceID int) string {
	return UidPidToken + stringutil.IntToString(userID) + ":" + model.DeviceIDToName(DeviceID)
}
