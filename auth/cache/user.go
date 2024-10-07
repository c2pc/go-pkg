package cache

import (
	"context"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/cache/cachekey"
	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/utils/cache"
	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
)

type IUserCache interface {
	cache.BatchDeleter
	GetUserInfo(ctx context.Context, userID int, fn func(ctx context.Context) (*model.User, error)) (userInfo *model.User, err error)
	DelUsersInfo(userIDs ...int) IUserCache
}

type UserCache struct {
	cache.BatchDeleter
	rdb          redis.UniversalClient
	rcClient     *rockscache.Client
	accessExpire time.Duration
}

func NewUserCache(rdb redis.UniversalClient, rcClient *rockscache.Client, batchHandler cache.BatchDeleter, accessExpire time.Duration) IUserCache {
	return &UserCache{
		BatchDeleter: batchHandler,
		rdb:          rdb,
		rcClient:     rcClient,
		accessExpire: accessExpire,
	}
}

func (u *UserCache) CloneUserCache() IUserCache {
	return &UserCache{
		BatchDeleter: u.BatchDeleter.Clone(),
		rdb:          u.rdb,
		accessExpire: u.accessExpire,
		rcClient:     u.rcClient,
	}
}

func (u *UserCache) GetUserInfo(ctx context.Context, userID int, fn func(ctx context.Context) (*model.User, error)) (userInfo *model.User, err error) {
	return cache.GetCache(ctx, u.rcClient, cachekey.GetUserInfoKey(userID), u.accessExpire, fn)
}

func (u *UserCache) DelUsersInfo(userIDs ...int) IUserCache {
	keys := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		keys = append(keys, cachekey.GetUserInfoKey(userID))
	}
	ch := u.CloneUserCache()
	ch.AddKeys(keys...)

	return ch
}
