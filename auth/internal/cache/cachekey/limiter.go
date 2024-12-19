package cachekey

const (
	USERNAME = "USERNAME:"
	USERIP   = "USERIP:"
)

func GetUsernameKey() string {
	return USERNAME
}

func GetUserIPKey() string {
	return USERIP
}
