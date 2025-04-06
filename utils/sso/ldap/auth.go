package ldap

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/go-ldap/ldap/v3"
)

var (
	ErrServerIsNotUnavailable = apperr.New("ldap_server_is_not_unavailable", apperr.WithTextTranslate(translator.Translate{translator.RU: "Сервер LDAP недоступен", translator.EN: "Server LDAP is unavailable"}), apperr.WithCode(code.Unavailable))
)

type AuthService interface {
	IsEnabled() bool
	CheckAuth(username, password string) error
}

type Config struct {
	Enabled    bool
	Addr       string
	BaseDN     string
	BaseFilter string
	LoginAttr  string
	Domain     string
}

type Auth struct {
	enabled    bool
	addr       string
	baseDN     string
	baseFilter string
	domain     string
	secured    bool
	loginAttr  string
}

func NewAuthService(cfg Config) (*Auth, error) {
	auth := new(Auth)
	auth.enabled = cfg.Enabled

	if auth.enabled {
		if cfg.LoginAttr == "" {
			return nil, errors.New("LDAP login attribute is required")
		}
		if cfg.BaseDN == "" {
			return nil, errors.New("LDAP base DN is required")
		}
		if cfg.Addr == "" {
			return nil, errors.New("LDAP addr is required")
		}
		if cfg.Domain == "" {
			return nil, errors.New("LDAP domain is required")
		}

		protoHostPort := strings.Split(cfg.Addr, "://")
		if len(protoHostPort) != 2 {
			err := fmt.Errorf("LDAP invalid URI: %s", cfg.Addr)
			return nil, err
		}

		if strings.ToUpper(protoHostPort[0]) == "LDAPS" {
			auth.secured = true
		} else {
			auth.secured = false
		}

		auth.domain = cfg.Domain
		auth.baseDN = cfg.BaseDN
		auth.baseFilter = cfg.BaseFilter
		auth.loginAttr = cfg.LoginAttr
		auth.addr = protoHostPort[1]
	}

	return auth, nil
}

func (a *Auth) IsEnabled() bool {
	return a.enabled
}

func (a *Auth) CheckAuth(username, password string) error {
	return a.bind(username, password)
}

func (a *Auth) bind(login, password string) error {
	var conn *ldap.Conn
	var err error

	if a.secured {
		conn, err = ldap.DialTLS("tcp", a.addr, &tls.Config{InsecureSkipVerify: true})
	} else {
		conn, err = ldap.Dial("tcp", a.addr)
	}
	if err != nil {
		return ErrServerIsNotUnavailable.WithError(err)
	}
	defer conn.Close()

	err = conn.Bind(fmt.Sprintf("%s@%s", login, a.domain), password)
	if err != nil {
		var e *ldap.Error
		if errors.As(err, &e) {
			if e.ResultCode == ldap.LDAPResultInvalidCredentials {
				return apperr.ErrUnauthenticated.WithErrorText("ldap invalid credentials")
			}
		}

		return apperr.ErrInternal.WithError(err)
	}

	entry, err := a.getDN(conn, login)
	if err != nil {
		return err
	}

	loginValue := entry.GetAttributeValue(a.loginAttr)
	if strings.ToLower(loginValue) != strings.ToLower(login) {
		return apperr.ErrUnauthenticated.WithErrorText(fmt.Sprintf("LDAP login attribute '%s' is not allowed (%s -> %s)", a.loginAttr, loginValue, login))
	}

	return nil
}

func (a *Auth) getDN(conn *ldap.Conn, login string) (*ldap.Entry, error) {
	var filter string
	if a.baseFilter != "" {
		filter = fmt.Sprintf("(&(%s=%s)(%s))", ldap.EscapeFilter(a.loginAttr), ldap.EscapeFilter(login), a.baseFilter)
	} else {
		filter = fmt.Sprintf("(%s=%s)", ldap.EscapeFilter(a.loginAttr), ldap.EscapeFilter(login))
	}

	searchReq := ldap.NewSearchRequest(
		a.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{},
		[]ldap.Control{},
	)
	result, err := conn.SearchWithPaging(searchReq, 100)
	if err != nil {
		return nil, apperr.ErrInternal.WithError(err)
	}

	if len(result.Entries) == 0 {
		return nil, apperr.ErrUnauthenticated.WithErrorText("no entries returned")
	}

	return result.Entries[0], nil
}
