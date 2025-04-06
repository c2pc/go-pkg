package ldap

import (
	"testing"
)

func TestName(t *testing.T) {
	auth, err := NewAuthService(Config{
		Enabled:    true,
		Addr:       "ldap://example.com",
		LoginAttr:  "sAMAccountName",
		BaseDN:     "DC=example,DC=com",
		BaseFilter: "(objectClass=Person)",
		Domain:     "example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = auth.CheckAuth("qwerty", "qwerty")
	if err != nil {
		t.Fatal(err)
	}
}
