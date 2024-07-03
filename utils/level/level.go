package level

type Level = string

const (
	PRODUCTION  Level = "production"
	DEVELOPMENT Level = "development"
	TEST        Level = "c2pc"
)

func Is(l string, level ...Level) bool {
	for _, lv := range level {
		if l == lv {
			return true
		}
	}
	return false
}
