package middleware

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/c2pc/go-pkg/v2/auth/cache"
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
	"github.com/gin-gonic/gin"
)

type IPermissionMiddleware interface {
	Can(c *gin.Context)
}

type PermissionMiddleware struct {
	debug                string
	userCache            cache.IUserCache
	permissionCache      cache.IPermissionCache
	userRepository       repository.IUserRepository
	permissionRepository repository.IPermissionRepository
}

func NewPermissionMiddleware(userCache cache.IUserCache, permissionCache cache.IPermissionCache, userRepository repository.IUserRepository,
	permissionRepository repository.IPermissionRepository, debug string) *PermissionMiddleware {
	return &PermissionMiddleware{
		debug:                debug,
		userCache:            userCache,
		permissionCache:      permissionCache,
		userRepository:       userRepository,
		permissionRepository: permissionRepository,
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
		return j.userRepository.GetUserWithPermissions(ctx, "id = ?", userID)
	})
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		c.Abort()
		return
	}

	permissions, err := j.permissionCache.GetPermissionList(ctx, func(ctx context.Context) ([]model.Permission, error) {
		return j.permissionRepository.List(ctx, &model2.Filter{}, ``)
	})
	if err != nil {
		response.Response(c, apperr.ErrInternal.WithError(err))
		c.Abort()
		return
	}

	if level.Is(j.debug, level.TEST) {
		logger.InfofLog(ctx, "PERMISSION", "[PERMISSIONS] %+v", permissions)
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

	if level.Is(j.debug, level.TEST) {
		logger.InfofLog(ctx, "PERMISSION", "[PATH] %s", perm)
	}

	if perm == "" {
		response.Response(c, apperr.ErrInternal.WithErrorText(c.FullPath()+" permission not found for "+c.Request.URL.Path))
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

	if level.Is(j.debug, level.TEST) {
		logger.InfofLog(ctx, "PERMISSION", "[PERMISSION] %+v", permission)
	}

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
		response.Response(c, apperr.ErrForbidden.WithErrorText("user haven't permission to access "+perm))
		c.Abort()
		return
	}

	c.Next()
	return
}
