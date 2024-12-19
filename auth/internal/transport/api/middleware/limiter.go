package middleware

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/cache"
	"github.com/c2pc/go-pkg/v2/auth/internal/cache/cachekey"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

type ConfigLimiter struct {
	MaxAttempts int
	TTL         time.Duration
}

type AuthMiddleware struct {
	debug string
	cfg   ConfigLimiter
	cache cache.ILimiterCache
}

func NewAuthLimiterMiddleware(cfg ConfigLimiter, cache cache.ILimiterCache, debug string) AuthMiddleware {
	return AuthMiddleware{cfg: cfg, cache: cache, debug: debug}
}

type AuthLimiter interface {
	LimiterMiddleware(c *gin.Context)
}

func (a *AuthMiddleware) limiter(c *gin.Context) {
	path := c.FullPath()
	var key1, key2 string

	if strings.Contains(path, "/auth/login") {
		cred, err := request2.BindJSON[request.AuthLoginRequest](c)

		if err != nil {
			response.Response(c, err)
			return
		}

		if cred.Login == "" {
			c.Next()
			return
		}

		key1 = cachekey.GetUsernameKey() + cred.Login
	}
	clientIP := c.ClientIP()

	key2 = cachekey.GetUserIPKey() + clientIP

	attempts1, err := a.cache.GetAttempts(c.Request.Context(), key1)

	attempts2, err := a.cache.GetAttempts(c.Request.Context(), key2)

	if level.Is(a.debug, level.TEST) && err != nil {
		logger.Warningf("[REDIS] error get attempts. %v", err)
	}

	if err != nil {
		return
	}

	if attempts1 >= a.cfg.MaxAttempts || attempts2 >= a.cfg.MaxAttempts {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "To many attempts.",
		})
		c.Abort()
		return
	}

	c.Next()

	statusCode := c.Writer.Status()
	if statusCode == http.StatusUnauthorized {
		_, err := a.cache.IncrAttempts(c.Request.Context(), key1, a.cfg.TTL)
		if level.Is(a.debug, level.TEST) && err != nil {
			logger.Warningf("[REDIS] error incr attempts.")
		}
		if err != nil {
			return
		}
		_, err = a.cache.IncrAttempts(c.Request.Context(), key2, a.cfg.TTL)
		if level.Is(a.debug, level.TEST) && err != nil {
			logger.Warningf("[REDIS] error incr attempts.")
		}
		if err != nil {
			return
		}
	}
}

func (a *AuthMiddleware) LimiterMiddleware(c *gin.Context) {
	a.limiter(c)
}
