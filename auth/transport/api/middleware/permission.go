package middleware

import (
	"context"
	"github.com/c2pc/go-pkg/v2/auth/cache"
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strings"
)

type IPermissionMiddleware interface {
	Can(c *gin.Context)
}

type PermissionMiddleware struct {
	userCache      cache.IUserCache
	userRepository repository.IUserRepository
}

func NewPermissionMiddleware(userCache cache.IUserCache, userRepository repository.IUserRepository) *PermissionMiddleware {
	return &PermissionMiddleware{
		userCache:      userCache,
		userRepository: userRepository,
	}
}

func (j *PermissionMiddleware) Can(c *gin.Context) {
	ctx := c.Request.Context()

	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		response.Response(c, apperr.ErrInternal.WithErrorText("error to get operation user id"))
		c.Abort()
		return
	}

	user, err := j.userCache.GetUserInfo(ctx, userID, func(ctx context.Context) (*model.User, error) {
		return j.userRepository.GetUserWithPermissions(ctx, "users.id = ?", userID)
	})
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		c.Abort()
		return
	}

	var perm string

	fullEls := strings.Split(c.FullPath(), "/")
	pathEls := strings.Split(c.Request.URL.Path, "/")
	path := strings.Join(stringutil.IntersectString(fullEls, pathEls), "/")

	re := regexp.MustCompile("^/api/v[0-9]+/(.*)$")
	match := re.FindStringSubmatch(path)
	if len(match) == 2 {
		perm = match[1]
	}

	if perm == "" {
		response.Response(c, apperr.ErrInternal.WithErrorText(c.FullPath()+" permission not found for "+c.Request.URL.Path))
	}

	isCan := func(perm string) bool {
		for _, role := range user.Roles {
			for _, rolePermission := range role.RolePermissions {
				if rolePermission.Permission.Name == perm {
					switch c.Request.Method {
					case http.MethodGet:
						if rolePermission.Read {
							return true
						}
					case http.MethodPost:
						fallthrough
					case http.MethodPut:
						fallthrough
					case http.MethodPatch:
						if rolePermission.Write {
							return true
						}
					case http.MethodDelete:
						if rolePermission.Exec {
							return true
						}
					}
				}
			}
		}

		return false
	}(perm)

	if !isCan {
		response.Response(c, apperr.ErrForbidden.WithErrorText("user haven't permission to access "+c.FullPath()))
		c.Abort()
		return
	}

	c.Next()
	return
}
