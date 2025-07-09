package middleware

import (
	"fmt"
	"testing"

	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
)

func TestGetClaimFromToken(t *testing.T) {
	claims, err := tokenverify.GetClaimFromToken(
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOjIsIkRldmljZUlEIjo0LCJleHAiOjE3NTIwODQzNDQsIm5iZiI6MTc1MjA4MzE0NCwiaWF0IjoxNzUyMDgzNDQ0fQ._AjfIg-p4qFmOaxrZyy8meLe4TZz2qLSLdTAHSmjZxI",
		tokenverify.Secret("123123"))
	if err != nil {

		fmt.Println(claims)
		t.Fatal(err)
	}

	fmt.Println(claims)
}
