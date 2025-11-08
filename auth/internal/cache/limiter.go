package cache

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
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

func NewLimiterCache(rdb redis.UniversalClient) *LimiterCache {
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
	script := redis.NewScript(`
		local current = redis.call("GET", KEYS[1])
		if not current then
			redis.call("SET", KEYS[1], 1, "EX", ARGV[1])
			return 1
		else
			local newval = redis.call("INCR", KEYS[1])
			return newval
		end
	`)

	result, err := script.Run(ctx, l.rdb, []string{key}, int(ttl.Seconds())).Result()
	if err != nil {
		return 0, err
	}

	attempts, ok := result.(int64)
	if !ok {
		return 0, errors.New("unexpected result type from redis script")
	}

	return int(attempts), nil
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
