package middleware

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/c2pc/go-pkg/v2/auth/internal/cache"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
	"github.com/gin-gonic/gin"
)

type IPermissionMiddleware interface {
	Can(c *gin.Context)
}

type PermissionMiddleware struct {
	cache        *cache.Cache
	repositories repository.Repositories
}

func NewPermissionMiddleware(cache *cache.Cache, repositories repository.Repositories) *PermissionMiddleware {
	return &PermissionMiddleware{
		cache:        cache,
		repositories: repositories,
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

	user, err := j.cache.UserCache.GetUserInfo(ctx, userID, func(ctx context.Context) (*model.User, error) {
		return j.repositories.UserRepository.GetUserWithPermissions(ctx, "id = ?", userID)
	})
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		c.Abort()
		return
	}

	permissions, err := j.cache.PermissionCache.GetPermissionList(ctx, func(ctx context.Context) ([]model.Permission, error) {
		return j.repositories.PermissionRepository.GetList(ctx)
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
		c.Abort()
		return
	}

	permission := func(perm string) *model.Permission {
		perms := strings.Split(perm, "/")
		for i := range perms {
			p2 := strings.Join(perms[0:len(perms)-i], "/")
			for _, p := range permissions {
				if p.Name == p2 {
					return &p
				}
			}
		}
		return nil
	}(perm)

	if permission == nil {
		c.Next()
		return
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
	}(permission.Name)

	if !isCan {
		c.Request = c.Request.WithContext(mcontext.WithOpActionContext(ctx, "Нет разрешения на доступ"))
		response.Response(c, apperr.ErrForbidden.WithErrorText("user haven't permission to access "+perm))
		c.Abort()
		return
	}

	c.Next()
	return
}
