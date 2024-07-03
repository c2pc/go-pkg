package tokenverify

import (
	"github.com/golang-jwt/jwt/v4"
	"testing"
)

var secret = "OpenIM_server"

func Test_ParseToken(t *testing.T) {
	claims1 := BuildClaims(123456, 1, 10)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims1)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	claim2, err := GetClaimFromToken(tokenString, secretFun())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(claim2)
}

func secretFun() jwt.Keyfunc {
	return func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	}
}
