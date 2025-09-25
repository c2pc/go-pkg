package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/c2pc/go-pkg/v2/auth/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/cache/cachekey"
	"github.com/c2pc/go-pkg/v2/auth/internal/transport/api/request"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	request2 "github.com/c2pc/go-pkg/v2/utils/request"
	response "github.com/c2pc/go-pkg/v2/utils/response/http"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin"
)

var (
	ErrToManyRequest = apperr.New("to_many_request", apperr.WithTextTranslate(
		translator.Translate{translator.RU: "Слишком много запросов", translator.EN: "Too many requests"}),
		apperr.WithCode(code.ResourceExhausted),
	)
)

type AuthMiddleware struct {
	cache *fx.CacheHolder
	cfg   *fx.LimiterHolder
}

func NewAuthLimiterMiddleware(cache *fx.CacheHolder, cfg *fx.LimiterHolder) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg, cache: cache}
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

	attempts, err := a.cache.Get().LimiterCache.GetAttempts(c.Request.Context(), key)
	if err != nil {
		return
	}

	if attempts >= a.cfg.Get().MaxAttempts {
		ttl, err := a.cache.Get().LimiterCache.GetTTL(c.Request.Context(), key)
		if err == nil {
			c.Request = c.Request.WithContext(mcontext.WithOpActionContext(c.Request.Context(), "Слишком много запросов"))
			c.Header("RateLimit-Limit", strconv.Itoa(a.cfg.Get().MaxAttempts))
			c.Header("RateLimit-Remaining", strconv.Itoa(attempts-a.cfg.Get().MaxAttempts))
			c.Header("RateLimit-Reset", strconv.FormatInt(int64(ttl.Seconds()), 10))
			response.Response(c, ErrToManyRequest)
			return
		}
	}

	c.Next()

	statusCode := c.Writer.Status()
	if statusCode == http.StatusUnauthorized {
		_, err = a.cache.Get().LimiterCache.IncrAttempts(c.Request.Context(), key, a.cfg.Get().TTL)
		if err != nil {
			return
		}
	}
}

func (a *AuthMiddleware) LimiterMiddleware(c *gin.Context) {
	a.limiter(c)
}
