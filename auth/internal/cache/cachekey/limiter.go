package cachekey

const (
	USERNAME = "USERNAME:"
	USERIP   = "USERIP:"
)

func GetUsernameKey() string {
	return ServiceName + USERNAME
}

func GetUserIPKey() string {
	return ServiceName + USERIP
}
