package cache

import (
	"context"
	"github.com/c2pc/go-pkg/v2/auth/cache/cachekey"
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/cache"
	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
	"time"
)

type IPermissionCache interface {
	cache.BatchDeleter
	GetPermissionList(ctx context.Context, fn func(ctx context.Context) ([]model.Permission, error)) (list []model.Permission, err error)
	DelPermissionList() IPermissionCache
}

type PermissionCache struct {
	cache.BatchDeleter
	rdb      redis.UniversalClient
	rcClient *rockscache.Client
}

func NewPermissionCache(rdb redis.UniversalClient, rcClient *rockscache.Client, batchHandler cache.BatchDeleter) IPermissionCache {
	c := &PermissionCache{
		BatchDeleter: batchHandler,
		rdb:          rdb,
		rcClient:     rcClient,
	}

	err := c.DelPermissionList().ChainExecDel(context.Background())
	if err != nil {
		panic(err)
	}

	return c
}

func (u *PermissionCache) ClonePermissionCache() IPermissionCache {
	return &PermissionCache{
		BatchDeleter: u.BatchDeleter.Clone(),
		rdb:          u.rdb,
		rcClient:     u.rcClient,
	}
}

func (u *PermissionCache) GetPermissionList(ctx context.Context, fn func(ctx context.Context) ([]model.Permission, error)) (list []model.Permission, err error) {
	return cache.GetCache(ctx, u.rcClient, cachekey.GetPermissionListKey(), 9*time.Hour, fn)
}

func (u *PermissionCache) DelPermissionList() IPermissionCache {
	ch := u.ClonePermissionCache()
	ch.AddKeys(cachekey.GetPermissionListKey())

	return ch
}
