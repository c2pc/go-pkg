package secret

import (
	"golang.org/x/crypto/bcrypt"
)

var HasherSecret Hasher

func init() {
	HasherSecret = new(Password)
}

type Hasher interface {
	HashString(str string) (string, error)
	HashMatchesString(hash, password string) bool
}

type Password struct {
}

func (p *Password) HashMatchesString(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (p *Password) HashString(str string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(str), 10)
	return string(bytes), err
}
