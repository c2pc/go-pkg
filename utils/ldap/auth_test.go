package ldap

import (
	"testing"
)

func TestName(t *testing.T) {
	auth, err := NewAuthService(Config{
		Enabled:    true,
		Addr:       "ldap://172.27.7.3:389",
		LoginAttr:  "sAMAccountName",
		BaseDN:     "DC=ts,DC=lab",
		BaseFilter: "(objectClass=Person)",
		Domain:     "ts.lab",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = auth.CheckAuth("админBeta12", "!23QweAsd")
	if err != nil {
		t.Fatal(err)
	}
}
