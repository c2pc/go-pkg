package secret

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

type Hasher interface {
	HashString(str string) string
	HashMatchesString(hash, password string) bool
}

func New(salt string) (*Password, error) {
	if salt == "" {
		return nil, errors.New("password salt is required")
	}

	return &Password{salt: salt}, nil
}

type Password struct {
	salt string
}

func (p *Password) HashMatchesString(hash, password string) bool {
	h := p.HashString(password)
	return p.HashMatches(hash, h)
}

func (p *Password) HashString(str string) string {
	sum := sha256.Sum256([]byte(str + p.salt))
	return fmt.Sprintf("%x", sum)
}

func (p *Password) HashMatches(h1, h2 string) bool {
	return h1 == h2
}
