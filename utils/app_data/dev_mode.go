package app_data

import (
	"os"
)

var devMode bool

func init() {
	devMode = os.Getenv("GO_DEV_MODE") == "true"
}

func IsDevMode() bool {
	return devMode
}
