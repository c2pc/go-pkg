package cache

import (
	"context"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/internal/cache/cachekey"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/utils/cache"
	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
)

type IUserCache interface {
	cache.BatchDeleter
	GetUserInfo(ctx context.Context, userID int64, fn func(ctx context.Context) (*model.User, error)) (userInfo *model.User, err error)
	DelUsersInfo(userIDs ...int64) IUserCache
}

type UserCache struct {
	cache.BatchDeleter
	rdb      redis.UniversalClient
	rcClient *rockscache.Client
}

func NewUserCache(rdb redis.UniversalClient, rcClient *rockscache.Client, batchHandler cache.BatchDeleter) *UserCache {
	return &UserCache{
		BatchDeleter: batchHandler,
		rdb:          rdb,
		rcClient:     rcClient,
	}
}

func (u *UserCache) CloneUserCache() IUserCache {
	return &UserCache{
		BatchDeleter: u.BatchDeleter.Clone(),
		rdb:          u.rdb,
		rcClient:     u.rcClient,
	}
}

func (u *UserCache) GetUserInfo(ctx context.Context, userID int64, fn func(ctx context.Context) (*model.User, error)) (userInfo *model.User, err error) {
	return cache.GetCache(ctx, u.rcClient, cachekey.GetUserInfoKey(userID), 1*time.Minute, fn)
}

func (u *UserCache) DelUsersInfo(userIDs ...int64) IUserCache {
	keys := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		keys = append(keys, cachekey.GetUserInfoKey(userID))
	}
	ch := u.CloneUserCache()
	ch.AddKeys(keys...)

	return ch
}
