package rbac

import (
	"context"
	"errors"
	"github.com/c2pc/go-pkg/apperr"
	"github.com/c2pc/go-pkg/jwt"
	"github.com/gin-gonic/gin"
)

var UnauthorizedAuthUser = errors.New("error to get authUser param from auth token")
var ForbiddenRole = errors.New("user does not have access")

type AuthUser struct {
	ID   int
	Role string
}

func User(ctx context.Context) (*AuthUser, error) {
	u, ok := ctx.Value(jwt.AuthUserKey).(*jwt.Claims)
	if !ok {
		return nil, apperr.ErrUnauthorized.WithError(UnauthorizedAuthUser)
	}
	return &AuthUser{
		ID:   u.Id,
		Role: u.Role,
	}, nil
}

func Can(roles ...string) func(c *gin.Context) {
	return func(c *gin.Context) {
		user, err := User(c.Request.Context())
		if err != nil {
			apperr.HTTPResponse(c, err)
			return
		}

		can := func() bool {
			for _, r := range roles {
				if user.Role == r {
					return true
				}
			}
			return false
		}()

		if can {
			c.Next()
		} else {
			apperr.HTTPResponse(c, apperr.ErrForbidden.WithError(ForbiddenRole))
			return
		}
	}
}
