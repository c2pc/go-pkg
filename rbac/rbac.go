package rbac

import (
	"context"
	"github.com/c2pc/go-pkg/apperr"
	"github.com/c2pc/go-pkg/apperr/utils/appErrors"
	"github.com/c2pc/go-pkg/apperr/utils/translate"
	"github.com/c2pc/go-pkg/apperr/x/httperr"
	"github.com/c2pc/go-pkg/jwt"
	"github.com/gin-gonic/gin"
)

var (
	ErrUnauthorizedMethod = apperr.New("auth",
		apperr.WithTitleTranslate(translate.Translate{translate.RU: "Попытка авторизации"}),
		apperr.WithContext("auth"),
	)

	ErrForbiddenMethod = apperr.New("auth",
		apperr.WithTitleTranslate(translate.Translate{translate.RU: "Попытка авторизации"}),
		apperr.WithContext("auth"),
	)

	ErrErrorToGetUserFromContext = apperr.New("error_to_get_user_from_context")
)

type AuthUser struct {
	ID           int
	Role         string
	Login        string
	DepartmentID *int
}

func User(ctx context.Context) (*AuthUser, error) {
	u, ok := ctx.Value(jwt.AuthUserKey).(*jwt.User)
	if !ok {
		return nil, ErrErrorToGetUserFromContext
	}

	return &AuthUser{
		ID:           u.Id,
		Role:         u.Role,
		Login:        u.Login,
		DepartmentID: u.DepartmentId,
	}, nil
}

func Can(roles ...string) func(c *gin.Context) {
	return func(c *gin.Context) {
		user, err := User(c.Request.Context())
		if err != nil {
			httperr.Response(c, ErrUnauthorizedMethod.WithError(appErrors.ErrUnauthenticated.WithError(err)))
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
			httperr.Response(c, ErrForbiddenMethod.WithError(appErrors.ErrForbidden.WithError(err)))
			return
		}
	}
}
