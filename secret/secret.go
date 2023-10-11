package secret

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

// New returns a password object
func New(salt string) (*Password, error) {
	if salt == "" {
		return nil, errors.New("password salt is empty")
	}
	return &Password{
		salt: salt,
	}, nil
}

// Password is our secret service implementation
type Password struct {
	salt string
}

func (p *Password) HashString(str string) string {
	sum := sha256.Sum256([]byte(str + p.salt))
	return fmt.Sprintf("%x", sum)
}

func (p *Password) HashMatchesCredentials(hash, password, login string) bool {
	h := p.HashString(password + login)
	return p.HashMatches(hash, h)
}

func (p *Password) HashMatches(h1, h2 string) bool {
	return h1 == h2
}
