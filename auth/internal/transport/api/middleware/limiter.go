package middleware

import (
	"github.com/c2pc/go-pkg/v2/auth/internal/cache"
	"github.com/c2pc/go-pkg/v2/auth/internal/cache/cachekey"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/level"

	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/c2pc/go-pkg/v2/utils/logger"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
)

var (
	ErrToManyRequest = apperr.New("to_many_request", apperr.WithTextTranslate(
		translator.Translate{translator.RU: "Много запросов", translator.EN: "To many request"}),
		apperr.WithCode(code.ResourceExhausted),
	)
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
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = 10
	}

	if cfg.TTL == 0 {
		cfg.TTL = time.Second
	}

	return AuthMiddleware{cfg: cfg, cache: cache, debug: debug}
}

type AuthLimiter interface {
	LimiterMiddleware(c *gin.Context)
}

func (a *AuthMiddleware) limiter(c *gin.Context) {
	path := c.FullPath()
	var key string
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

		key = cachekey.GetUsernameKey() + cred.Login
	} else {
		clientIP := c.ClientIP()
		key = cachekey.GetUserIPKey() + clientIP
	}

	attempts, err := a.cache.GetAttempts(c.Request.Context(), key)

	if level.Is(a.debug, level.TEST) && err != nil {
		logger.Warningf("[REDIS] error get attempts - %v", err)
	}

	if err != nil {
		return
	}

	if attempts >= a.cfg.MaxAttempts {
		ttl, err := a.cache.GetTTL(c.Request.Context(), key)
		if err != nil {
			if level.Is(a.debug, level.TEST) {
				logger.Warningf("[REDIS] error get TTL - %v", err)
			}
		} else {
			c.Header("RateLimit-Limit", strconv.Itoa(a.cfg.MaxAttempts))
			c.Header("RateLimit-Remaining", strconv.Itoa(attempts-a.cfg.MaxAttempts))
			c.Header("RateLimit-Reset", strconv.FormatInt(int64(ttl.Seconds()), 10))
			response.Response(c, ErrToManyRequest)
			return
		}
	}

	c.Next()

	statusCode := c.Writer.Status()
	if statusCode == http.StatusUnauthorized {
		_, err = a.cache.IncrAttempts(c.Request.Context(), key, a.cfg.TTL)
		if level.Is(a.debug, level.TEST) && err != nil {
			logger.Warningf("[REDIS] error incr attempts - %v", err)
		}
		if err != nil {
			return
		}
	}
}

func (a *AuthMiddleware) LimiterMiddleware(c *gin.Context) {
	a.limiter(c)
}
