package middleware

import (
	"testing"

	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
)

func TestGetClaimFromToken(t *testing.T) {
	_, err := tokenverify.GetClaimFromToken(
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySUQiOjIsIkRldmljZUlEIjo0LCJleHAiOjE3NTIwODQzNDQsIm5iZiI6MTc1MjA4MzE0NCwiaWF0IjoxNzUyMDgzNDQ0fQ._AjfIg-p4qFmOaxrZyy8meLe4TZz2qLSLdTAHSmjZxI",
		tokenverify.Secret("123123"))
	if err != nil {
		t.Fatal(err)
	}
}
