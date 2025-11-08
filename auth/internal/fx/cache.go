package fx

import (
	"sync/atomic"
	"time"

	cache2 "github.com/c2pc/go-pkg/v2/auth/internal/cache"
	"github.com/c2pc/go-pkg/v2/utils/cache"
	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
)

type CacheHolder struct {
	rcClient     *rockscache.Client
	batchHandler *cache.BatchDeleterRedis
	v            atomic.Value
}

func NewCacheHolder(rdb redis.UniversalClient, accessTokenTTL time.Duration) *CacheHolder {
	rcClient := rockscache.NewClient(rdb, cache.GetRocksCacheOptions())
	batchHandler := cache.NewBatchDeleterRedis(rdb, cache.GetRocksCacheOptions())

	h := &CacheHolder{
		rcClient:     rcClient,
		batchHandler: batchHandler,
	}

	h.Reload(rdb, accessTokenTTL)
	return h
}

func (h *CacheHolder) Get() *cache2.Cache { return h.v.Load().(*cache2.Cache) }

func (h *CacheHolder) Reload(rdb redis.UniversalClient, accessTokenTTL time.Duration) {
	svc := cache2.NewCache(rdb, h.rcClient, h.batchHandler, accessTokenTTL)
	h.v.Store(svc)
	return
}
