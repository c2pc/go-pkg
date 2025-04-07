package secret

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Hasher interface {
	HashString(str string) (string, error)
	HashMatchesString(hash, password string) bool
}

func New(salt string, useBcrypt bool) (*Password, error) {
	if !useBcrypt {
		if salt == "" {
			return nil, errors.New("password salt is required")
		}
	}

	return &Password{salt: salt, useBcrypt: useBcrypt}, nil
}

type Password struct {
	salt      string
	useBcrypt bool
}

func (p *Password) HashMatchesString(hash, password string) bool {
	if p.useBcrypt {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		return err == nil
	}

	return hash == p.hashStringWithSalt(password)
}

func (p *Password) HashString(str string) (string, error) {
	if p.useBcrypt {
		bytes, err := bcrypt.GenerateFromPassword([]byte(str), 14)
		return string(bytes), err
	}

	sum := sha256.Sum256([]byte(str + p.salt))
	return fmt.Sprintf("%x", sum), nil
}

func (p *Password) hashStringWithSalt(str string) string {
	sum := sha256.Sum256([]byte(str + p.salt))
	return fmt.Sprintf("%x", sum)
}
