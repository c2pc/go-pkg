package cache

import (
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type ILimiterCache interface {
	GetAttempts(ctx context.Context, key string) (int, error)
	IncrAttempts(ctx context.Context, key string, ttl time.Duration) (int, error)
	SetTTL(ctx context.Context, key string, ttl time.Duration) error
	ResetAttempts(ctx context.Context, key string) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)
}

type LimiterCache struct {
	rdb redis.UniversalClient
}

func NewLimiterCache(rdb redis.UniversalClient) ILimiterCache {
	return &LimiterCache{
		rdb: rdb,
	}
}

func (l LimiterCache) GetAttempts(ctx context.Context, key string) (int, error) {
	val, err := l.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}

	if err != nil {
		return 0, err
	}

	attempts, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}

	return attempts, nil
}

func (l LimiterCache) IncrAttempts(ctx context.Context, key string, ttl time.Duration) (int, error) {
	exists, err := l.rdb.Exists(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if exists == 0 {
		err = l.rdb.Set(ctx, key, 1, ttl).Err()
		if err != nil {
			return 0, err
		}
		return 1, nil
	} else {
		attempts, err := l.rdb.Incr(ctx, key).Result()
		if err != nil {
			return 0, err
		}

		err = l.rdb.Expire(ctx, key, ttl).Err()
		if err != nil {
			return 0, err
		}

		return int(attempts), nil
	}
}

func (l LimiterCache) SetTTL(ctx context.Context, key string, ttl time.Duration) error {
	return l.rdb.Expire(ctx, key, ttl).Err()
}

func (l LimiterCache) ResetAttempts(ctx context.Context, key string) error {
	return l.rdb.Del(ctx, key).Err()
}

func (l LimiterCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return l.rdb.TTL(ctx, key).Result()
}
