package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/redis/go-redis/v9"
)

// RedisClient defines the configuration parameters for a Redis client, including
// options for both single-node and cluster mode connections.
type RedisClient struct {
	ClusterMode bool     // Whether to use Redis in cluster mode.
	Address     []string // List of Redis server addresses (host:port).
	Username    string   // Username for Redis authentication (Redis 6 ACL).
	Password    string   // Password for Redis authentication.
	MaxRetry    int      // Maximum number of retries for a command.
	DB          int      // Database number to connect to, for non-cluster mode.
	PoolSize    int      // Number of connections to pool.
}

func NewRedisClient(ctx context.Context, config *RedisClient, debug string) (redis.UniversalClient, error) {
	if len(config.Address) == 0 {
		return nil, errors.New("redis address is empty")
	}
	var cli redis.UniversalClient
	if config.ClusterMode || len(config.Address) > 1 {
		opt := &redis.ClusterOptions{
			Addrs:      config.Address,
			Username:   config.Username,
			Password:   config.Password,
			PoolSize:   config.PoolSize,
			MaxRetries: config.MaxRetry,
		}
		cli = redis.NewClusterClient(opt)
	} else {
		opt := &redis.Options{
			Addr:       config.Address[0],
			Username:   config.Username,
			Password:   config.Password,
			DB:         config.DB,
			PoolSize:   config.PoolSize,
			MaxRetries: config.MaxRetry,
		}
		cli = redis.NewClient(opt)
	}

	if level.Is(debug, level.TEST) {
		redis.SetLogger(defaultLogger())
	}

	if err := cli.Ping(ctx).Err(); err != nil {
		return nil, errors.New(fmt.Sprintf("Redis Ping failed %v", err))
	}

	return cli, nil
}
