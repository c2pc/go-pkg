package cache

import (
	"github.com/c2pc/go-pkg/v2/utils/cache"
	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	UserCache *UserCache
}

func NewCache(rdb redis.UniversalClient) *Cache {
	rcClient := rockscache.NewClient(rdb, cache.GetRocksCacheOptions())
	batchHandler := cache.NewBatchDeleterRedis(rdb, cache.GetRocksCacheOptions())

	return &Cache{
		UserCache: NewUserCache(rdb, rcClient, batchHandler),
	}
}
