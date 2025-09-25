package ldap

import (
	"crypto/tls"
	"errors"
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
	CheckAuth(domain, username, password string) error
}

type Config struct {
	Addrs  []string
	Domain string
}

type Auth struct {
	enabled bool
	cfg     map[string]Config
}

func NewAuthService(enabled bool, cfg map[string]Config) *Auth {
	return &Auth{
		enabled: enabled,
		cfg:     cfg,
	}
}

func (a *Auth) IsEnabled() bool {
	return a.enabled
}

func (a *Auth) CheckAuth(domain, username, password string) error {
	return a.bind(domain, username, password)
}

func (a *Auth) bind(domain, login, password string) error {
	var conn *ldap.Conn
	var err error

	var ld Config
	if domain == "" {
		for _, l := range a.cfg {
			ld = l
			break
		}
		login = login + "@" + ld.Domain
	} else {
		l, ok := a.cfg[domain]
		if !ok {
			return ErrServerIsNotUnavailable.WithErrorText("domain not found in config")
		}
		ld = l
	}

	for _, server := range ld.Addrs {
		var opts []ldap.DialOpt
		if strings.HasPrefix(server, "ldaps://") {
			opts = []ldap.DialOpt{
				ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
			}
		}

		conn, err = ldap.DialURL(server, opts...)
		if err == nil {
			break
		}
	}

	if conn == nil {
		return ErrServerIsNotUnavailable.WithError(err)
	} else {
		defer conn.Close()
	}

	err = conn.Bind(login, password)
	if err != nil {
		var e *ldap.Error
		if errors.As(err, &e) {
			if e.ResultCode == ldap.LDAPResultInvalidCredentials {
				return apperr.ErrUnauthenticated.WithErrorText("ldap invalid credentials")
			}
		}

		return apperr.ErrInternal.WithError(err)
	}

	return nil
}
