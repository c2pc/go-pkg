package ldap

import (
	"testing"
)

func TestName(t *testing.T) {

	auth, err := NewAuthService(Config{
		Enabled:    true,
		Addr:       "ldap://127.0.0.1:389",
		LoginAttr:  "sAMAccountName",
		BaseDN:     "OU=ru,DC=example,DC=com",
		BaseFilter: "(objectClass=Person)",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = auth.CheckAuth("admin", "admin")
	if err != nil {
		t.Fatal(err)
	}
}
