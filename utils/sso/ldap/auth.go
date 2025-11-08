package ldap

import (
	"crypto/tls"
	"errors"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/go-ldap/ldap/v3"
)

var (
	ErrServerIsNotUnavailable = apperr.New("ldap_server_is_not_unavailable", apperr.WithTextTranslate(translator.Translate{translator.RU: "Сервер LDAP недоступен", translator.EN: "Server LDAP is unavailable"}), apperr.WithCode(code.Unavailable))
	ErrDomainNotFound         = apperr.New("ldap_domain_not_found", apperr.WithTextTranslate(translator.Translate{translator.RU: "Доменная аутентификация не настроена", translator.EN: "The domain auth is not configured"}), apperr.WithCode(code.Unavailable))
)

type AuthService interface {
	IsEnabled() bool
	CheckAuth(domain, username, password string) error
}

type Config struct {
	Domain  string
	Secured bool
	Md5hash bool
	Addrs   []string
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

func (a *Auth) GetConfigs() map[string]Config {
	return a.cfg
}

func (a *Auth) CheckAuth(domain, username, password string) error {
	return a.bind(domain, username, password)
}

func (a *Auth) bind(domain, login, password string) error {
	ld, ok := a.cfg[domain]
	if !ok {
		return ErrDomainNotFound
	}

	var host string
	var conn *ldap.Conn
	var err error
	for _, u := range ld.Addrs {
		host = u

		var (
			opts []ldap.DialOpt
			url  string
		)
		if ld.Secured {
			opts = []ldap.DialOpt{
				ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
			}
			url = "ldaps://" + u
		} else {
			url = "ldap://" + u
		}

		conn, err = ldap.DialURL(url, opts...)
		if err == nil {
			break
		}
	}

	if conn == nil {
		return ErrServerIsNotUnavailable.WithError(err)
	} else {
		defer conn.Close()
	}

	if ld.Md5hash {
		err = conn.MD5Bind(host, login, password)
	} else {
		err = conn.Bind(login, password)
	}

	if err != nil {
		var e *ldap.Error
		if errors.As(err, &e) {
			if e.ResultCode == ldap.LDAPResultInvalidCredentials {
				return apperr.ErrUnauthenticated.WithErrorText("Неправильный логин или пароль")
			}
		}

		return apperr.ErrInternal.WithError(err)
	}

	return nil
}
