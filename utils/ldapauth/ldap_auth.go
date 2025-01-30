package ldapauth

import (
	"github.com/c2pc/go-pkg/v2/utils/ldapauth/internal/config"
	"github.com/c2pc/go-pkg/v2/utils/ldapauth/internal/service"
)

type AuthService interface {
	Login(username, password string) (*TokenResponse, error)
	Refresh(refreshToken string) (*TokenResponse, error)
}

type Config = config.Config
type TokenResponse = service.TokenResponse

func NewAuthService(cfg Config) AuthService {
	return service.NewAuthService(cfg)
}
