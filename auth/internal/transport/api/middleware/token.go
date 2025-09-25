package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/c2pc/go-pkg/v2/auth/fx"
	model3 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
	"github.com/gin-gonic/gin"
)

const authorizationHeader = "Authorization"
const authorizationQuery = "token"

type ITokenMiddleware interface {
	Authenticate(c *gin.Context)
}

type TokenMiddleware struct {
	cache        *fx.CacheHolder
	repositories repository.Repositories
	authHolder   *fx.AuthHolder
}

func NewTokenMiddleware(cache *fx.CacheHolder, repositories repository.Repositories, authHolder *fx.AuthHolder) *TokenMiddleware {
	return &TokenMiddleware{
		cache:        cache,
		repositories: repositories,
		authHolder:   authHolder,
	}
}

func (j *TokenMiddleware) Authenticate(c *gin.Context) {
	ctx := c.Request.Context()

	tokensString, err := j.parseAuthHeader(c)
	if err != nil {
		http.Response(c, apperr.ErrUnauthenticated.WithError(err))
		return
	}

	claims, err := tokenverify.GetClaimFromToken(tokensString, tokenverify.Secret(j.authHolder.Get().AccessSecret))
	if claims != nil {
		ctx = mcontext.WithOpUserIDContext(ctx, claims.UserID)
		ctx = mcontext.WithOpDeviceIDContext(ctx, claims.DeviceID)
		c.Request = c.Request.WithContext(ctx)
	}
	if claims == nil || err != nil {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(ctx, "Неверный токен доступа"))
		http.Response(c, apperr.ErrUnauthenticated.WithError(err))
		return
	}

	m, err := j.cache.Get().TokenCache.GetTokensWithoutError(ctx, claims.UserID, claims.DeviceID)
	if err != nil {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(ctx, "Неверный токен доступа"))
		http.Response(c, apperr.ErrInternal.WithError(err))
		return
	}
	if len(m) == 0 {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(ctx, "Неверный токен доступа"))
		http.Response(c, apperr.ErrUnauthenticated.WithError(tokenverify.ErrTokenNotExist))
		return
	}

	user, err := j.cache.Get().UserCache.GetUserInfo(ctx, claims.UserID, func(ctx context.Context) (*model3.User, error) {
		return j.repositories.UserRepository.GetUserWithPermissions(ctx, "id = ?", claims.UserID)
	})
	if err != nil {
		http.Response(c, apperr.ErrInternal.WithError(err))
		return
	}

	var roleName string
	var logDisabled bool
	for _, role := range user.Roles {
		if role.Name == model3.Broker {
			roleName = role.Name
			logDisabled = role.LogDisabled
			break
		} else if role.Name == model3.SuperAdmin {
			roleName = role.Name
			break
		}
	}

	if roleName != "" {
		ctx = mcontext.WithOpUserRoleContext(ctx, roleName)
		ctx = context.WithValue(ctx, "log_disabled", logDisabled)
	}
	c.Request = c.Request.WithContext(ctx)

	if v, ok := m[tokensString]; ok {
		switch v {
		case constant.NormalToken:
			c.Next()
			return
		case constant.KickedToken:
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(ctx, "Неверный токен доступа"))
			http.Response(c, apperr.ErrUnauthenticated.WithError(tokenverify.ErrTokenKicked))
			c.Abort()
			return
		default:
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(ctx, "Неверный токен доступа"))
			http.Response(c, apperr.ErrUnauthenticated.WithError(tokenverify.ErrTokenUnknown))
			c.Abort()
			return
		}
	}

	c.Request = c.Request.WithContext(mcontext.WithOpActionContext(ctx, "Неверный токен доступа"))
	http.Response(c, apperr.ErrUnauthenticated.WithError(tokenverify.ErrTokenNotExist))
	c.Abort()
	return
}

func (j *TokenMiddleware) parseAuthHeader(c *gin.Context) (string, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		token, ok := c.GetQuery(authorizationQuery)
		if ok && token != "" {
			return token, nil
		} else {
			return "", errors.New("empty auth header")
		}
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", errors.New("invalid auth header")
	}
	if len(headerParts[1]) == 0 {
		return "", errors.New("token is empty")
	}

	return headerParts[1], nil
}
