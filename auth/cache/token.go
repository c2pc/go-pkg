package cache

import (
	"context"
	"github.com/c2pc/go-pkg/v2/auth/cache/cachekey"
	"github.com/c2pc/go-pkg/v2/utils/stringutil"
	"github.com/redis/go-redis/v9"
	"time"
)

type ITokenCache interface {
	SetTokenFlag(ctx context.Context, userID int, DeviceID int, token string, flag int) error
	SetTokenFlagEx(ctx context.Context, userID int, DeviceID int, token string, flag int) error
	GetTokensWithoutError(ctx context.Context, userID int, DeviceID int) (map[string]int, error)
	SetTokenMapByUidPid(ctx context.Context, userID int, DeviceID int, m map[string]int) error
	DeleteTokenByUidPid(ctx context.Context, userID int, DeviceID int, fields []string) error
}

type TokenCache struct {
	rdb          redis.UniversalClient
	accessExpire time.Duration
}

func NewTokenCache(rdb redis.UniversalClient, accessExpire time.Duration) ITokenCache {
	return &TokenCache{
		rdb:          rdb,
		accessExpire: accessExpire,
	}
}

func (c *TokenCache) SetTokenFlag(ctx context.Context, userID int, DeviceID int, token string, flag int) error {
	return c.rdb.HSet(ctx, cachekey.GetTokenKey(userID, DeviceID), token, flag).Err()
}

// SetTokenFlagEx set token and flag with expire time
func (c *TokenCache) SetTokenFlagEx(ctx context.Context, userID int, DeviceID int, token string, flag int) error {
	key := cachekey.GetTokenKey(userID, DeviceID)
	if err := c.rdb.HSet(ctx, key, token, flag).Err(); err != nil {
		return err
	}
	if err := c.rdb.Expire(ctx, key, c.accessExpire).Err(); err != nil {
		return err
	}
	return nil
}

func (c *TokenCache) GetTokensWithoutError(ctx context.Context, userID int, DeviceID int) (map[string]int, error) {
	m, err := c.rdb.HGetAll(ctx, cachekey.GetTokenKey(userID, DeviceID)).Result()
	if err != nil {
		return nil, err
	}
	mm := make(map[string]int)
	for k, v := range m {
		mm[k] = stringutil.StringToInt(v)
	}

	return mm, nil
}

func (c *TokenCache) SetTokenMapByUidPid(ctx context.Context, userID int, DeviceID int, m map[string]int) error {
	mm := make(map[string]any)
	for k, v := range m {
		mm[k] = v
	}
	return c.rdb.HSet(ctx, cachekey.GetTokenKey(userID, DeviceID), mm).Err()
}

func (c *TokenCache) DeleteTokenByUidPid(ctx context.Context, userID int, DeviceID int, fields []string) error {
	return c.rdb.HDel(ctx, cachekey.GetTokenKey(userID, DeviceID), fields...).Err()
}
