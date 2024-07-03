package middleware

import (
	"errors"
	"github.com/c2pc/go-pkg/v2/auth/cache"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"github.com/c2pc/go-pkg/v2/utils/response/httperr"
	"github.com/c2pc/go-pkg/v2/utils/tokenverify"
	"github.com/gin-gonic/gin"
	"strings"
)

const authorizationHeader = "Authorization"

type ITokenMiddleware interface {
	Authenticate(c *gin.Context)
}

type TokenMiddleware struct {
	authCache cache.ITokenCache
	secret    string
}

func NewTokenMiddleware(authCache cache.ITokenCache, secret string) *TokenMiddleware {
	return &TokenMiddleware{
		authCache: authCache,
		secret:    secret,
	}
}

func (j *TokenMiddleware) Authenticate(c *gin.Context) {
	ctx := c.Request.Context()

	tokensString, err := j.parseAuthHeader(c)
	if err != nil {
		httperr.Response(c, apperr.ErrUnauthenticated.WithError(err))
		return
	}

	claims, err := tokenverify.GetClaimFromToken(tokensString, tokenverify.Secret(j.secret))
	if err != nil {
		httperr.Response(c, apperr.ErrUnauthenticated.WithError(err))
		return
	}

	m, err := j.authCache.GetTokensWithoutError(ctx, claims.UserID, claims.DeviceID)
	if err != nil {
		httperr.Response(c, apperr.ErrUnauthenticated.WithError(err))
		return
	}
	if len(m) == 0 {
		httperr.Response(c, apperr.ErrUnauthenticated.WithError(tokenverify.ErrTokenNotExist))
	}

	if v, ok := m[tokensString]; ok {
		switch v {
		case constant.NormalToken:
			ctx = mcontext.WithOpUserIDContext(ctx, claims.UserID)
			ctx = mcontext.WithOpDeviceIDContext(ctx, claims.DeviceID)

			c.Request = c.Request.WithContext(ctx)
			c.Next()
		case constant.KickedToken:
			httperr.Response(c, apperr.ErrUnauthenticated.WithError(tokenverify.ErrTokenKicked))
			c.Abort()
			return
		default:
			httperr.Response(c, apperr.ErrUnauthenticated.WithError(tokenverify.ErrTokenUnknown))
			c.Abort()
			return
		}
	}

	httperr.Response(c, apperr.ErrUnauthenticated.WithError(tokenverify.ErrTokenNotExist))
	c.Abort()
	return
}

func (j *TokenMiddleware) parseAuthHeader(c *gin.Context) (string, error) {
	header := c.GetHeader(authorizationHeader)
	if header == "" {
		return "", errors.New("empty auth header")
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
