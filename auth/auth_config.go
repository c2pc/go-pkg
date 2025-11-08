package auth

import (
	"context"
	"encoding/json"
	"time"

	authConf "github.com/c2pc/go-pkg/v2/auth/internal/configurator"
	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/service"
	"github.com/c2pc/go-pkg/v2/utils/sso/ldap"
	"github.com/c2pc/go-pkg/v2/utils/sso/oidc"
	"github.com/c2pc/go-pkg/v2/utils/sso/saml"
)

func getAuthCfg(ctx context.Context, configService service.ConfigService, authConfig *authConf.AConfigurator) (authConf.ACfg, error) {
	cfg, err := configService.SetConfig(ctx, model.AuthConfigKey, authConfig)
	if err != nil {
		return authConf.ACfg{}, err
	}

	authCfg, err := authConfig.Unmarshal(cfg)
	if err != nil {
		return authConf.ACfg{}, err
	}

	return authCfg.(authConf.ACfg), nil
}

func getLdapConfig(c []authConf.ACfgLDAP) map[string]ldap.Config {
	cfg := make(map[string]ldap.Config)
	for _, conf := range c {
		cfg[conf.Domain] = ldap.Config{
			Addrs:   conf.Addrs,
			Domain:  conf.Domain,
			Secured: conf.Secured,
			Md5hash: conf.Md5hash,
		}
	}
	return cfg
}

func getOIDCConfig(enabled bool, cfg *authConf.ACfgOIDC, desc string) oidc.Config {
	if cfg == nil {
		return oidc.Config{
			Enabled: false,
		}
	}

	return oidc.Config{
		Enabled:           enabled,
		Description:       desc,
		ConfigURL:         cfg.ConfigURL,
		ClientID:          cfg.ClientID,
		ClientSecret:      string(cfg.ClientSecret),
		RootURL:           cfg.RootURL,
		LoginAttr:         cfg.LoginAttr,
		ValidRedirectURLs: cfg.ValidRedirectURLs,
	}
}

func getSAMLConfig(enabled bool, cfg *authConf.ACfgSAML, desc string) saml.Config {
	if cfg == nil {
		return saml.Config{
			Enabled: false,
		}
	}

	return saml.Config{
		Enabled:           enabled,
		Description:       desc,
		MetaDataFile:      cfg.MetaDataFile,
		CertFile:          cfg.CertFile,
		KeyFile:           cfg.KeyFile,
		RootURL:           cfg.RootURL,
		LoginAttr:         cfg.LoginAttr,
		ValidRedirectURLs: cfg.ValidRedirectURLs,
	}
}

func getLimiterConfig(cfg authConf.ACfgLimiter) fx.ConfigLimiter {
	return fx.ConfigLimiter{
		MaxAttempts: cfg.MaxAttempts,
		TTL:         time.Duration(cfg.TTL) * time.Minute,
		BlockingTTL: time.Duration(cfg.BlockingTTL) * time.Minute,
	}
}

func (a *Auth) authConfigWatcher(ctx context.Context, config *authConf.AConfigurator) {
	for {
		select {
		case <-ctx.Done():
			return
		case newCfg := <-config.Watch():
			oldCfg := a.authCfg

			accessTokenTTL := time.Duration(newCfg.AccessTokenTTL) * time.Minute
			refreshExpire := time.Duration(newCfg.RefreshTokenTTL) * time.Minute
			accessSecret := newCfg.AccessKey

			if newCfg.AccessTokenTTL != oldCfg.AccessTokenTTL || newCfg.RefreshTokenTTL != oldCfg.RefreshTokenTTL || string(newCfg.AccessKey) != string(oldCfg.AccessKey) {
				a.cacheHolder.Reload(a.rdb, accessTokenTTL)
				a.authHolder.Reload(accessTokenTTL, refreshExpire, string(accessSecret))
			}

			if newCfg.Limiter.MaxAttempts != oldCfg.Limiter.MaxAttempts || newCfg.Limiter.TTL != oldCfg.Limiter.TTL || newCfg.Limiter.BlockingTTL != oldCfg.Limiter.BlockingTTL {
				a.limiterHolder.Reload(getLimiterConfig(newCfg.Limiter))
			}

			m1, _ := json.Marshal(newCfg.LDAP)
			m2, _ := json.Marshal(oldCfg.LDAP)

			if string(m1) != string(m2) || len(newCfg.LDAP) != len(oldCfg.LDAP) {
				a.ldapHolder.Reload(len(newCfg.LDAP) != 0, getLdapConfig(newCfg.LDAP))
			}

			_ = a.oidcHolder.Reload(ctx, getOIDCConfig(newCfg.SSO.Enabled == "oidc", newCfg.SSO.OIDC, newCfg.SSO.Description))
			_ = a.samlHolder.Reload(ctx, getSAMLConfig(newCfg.SSO.Enabled == "saml", newCfg.SSO.SAML, newCfg.SSO.Description))

			a.authCfg = newCfg
		}
	}
}
