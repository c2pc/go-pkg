package secret

// Service is the interface to our secret service
type Service interface {
	HashString(str string) string
	HashMatchesCredentials(hash, password, login string) bool
	HashMatches(h1, h2 string) bool
}
