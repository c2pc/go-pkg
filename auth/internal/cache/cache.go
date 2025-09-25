package cache

import (
	"time"

	"github.com/c2pc/go-pkg/v2/utils/cache"
	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	LimiterCache    *LimiterCache
	PermissionCache *PermissionCache
	TokenCache      *TokenCache
	UserCache       *UserCache
}

func NewCache(rdb redis.UniversalClient, rcClient *rockscache.Client, batchHandler *cache.BatchDeleterRedis, accessExpire time.Duration) *Cache {
	return &Cache{
		LimiterCache:    NewLimiterCache(rdb),
		PermissionCache: NewPermissionCache(rdb, rcClient, batchHandler),
		TokenCache:      NewTokenCache(rdb, accessExpire),
		UserCache:       NewUserCache(rdb, rcClient, batchHandler),
	}
}
