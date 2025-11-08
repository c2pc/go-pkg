package app_data

import (
	"os"
)

var LogLevel int8
var devMode bool

func init() {
	devMode = os.Getenv("GO_DEV_MODE") == "true"
}

func IsDevMode() bool {
	return devMode || LogLevel == 0
}
