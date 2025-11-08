package cachekey

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
)

const (
	UidPidToken = "TOKEN_STATUS:"
)

func GetTokenKey(userID int64, DeviceID int) string {
	return ServiceName + UidPidToken + stringutil.Int64ToString(userID) + ":" + model.DeviceIDToName(DeviceID)
}
