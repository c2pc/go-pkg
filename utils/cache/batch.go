package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/datautil"
	"github.com/dtm-labs/rockscache"
	"github.com/redis/go-redis/v9"
	"time"
)

// BatchDeleter интерфейс для выполнения пакетного удаления ключей из кэша
type BatchDeleter interface {
	// ChainExecDel метод используется для цепочечных вызовов и должен вызывать Clone для предотвращения загрязнения памяти.
	ChainExecDel(ctx context.Context) error
	// ExecDelWithKeys метод принимает ключи для удаления.
	ExecDelWithKeys(ctx context.Context, keys []string) error
	// Clone метод создает копию BatchDeleter, чтобы избежать модификации исходного объекта.
	Clone() BatchDeleter
	// AddKeys метод добавляет ключи для удаления.
	AddKeys(keys ...string)
}

// BatchDeleterRedis конкретная реализация интерфейса BatchDeleter, основанная на Redis и RocksCache.
type BatchDeleterRedis struct {
	redisClient redis.UniversalClient
	keys        []string
	rocksClient *rockscache.Client
}

// NewBatchDeleterRedis создает новый экземпляр BatchDeleterRedis.
func NewBatchDeleterRedis(redisClient redis.UniversalClient, options rockscache.Options) *BatchDeleterRedis {
	return &BatchDeleterRedis{
		redisClient: redisClient,
		rocksClient: rockscache.NewClient(redisClient, options),
	}
}

// ExecDelWithKeys напрямую принимает ключи для пакетного удаления и публикует информацию об удалении.
func (c *BatchDeleterRedis) ExecDelWithKeys(ctx context.Context, keys []string) error {
	// Убираем дубликаты из ключей
	distinctKeys := datautil.Distinct(keys)
	return c.execDel(ctx, distinctKeys)
}

// ChainExecDel используется для цепочечных вызовов для пакетного удаления. Необходимо вызвать Clone для предотвращения загрязнения памяти.
func (c *BatchDeleterRedis) ChainExecDel(ctx context.Context) error {
	// Убираем дубликаты из ключей
	distinctKeys := datautil.Distinct(c.keys)
	return c.execDel(ctx, distinctKeys)
}

// execDel выполняет пакетное удаление и публикует удаленные ключи для обновления информации о локальном кэше других узлов.
func (c *BatchDeleterRedis) execDel(ctx context.Context, keys []string) error {
	if len(keys) > 0 {
		// Пакетное удаление ключей
		err := ProcessKeysBySlot(ctx, c.redisClient, keys, func(ctx context.Context, slot int64, keys []string) error {
			return c.rocksClient.TagAsDeletedBatch2(ctx, keys)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Clone создает копию BatchDeleterRedis для цепочечных вызовов, чтобы предотвратить загрязнение памяти.
func (c *BatchDeleterRedis) Clone() BatchDeleter {
	return &BatchDeleterRedis{
		redisClient: c.redisClient,
		keys:        c.keys,
		rocksClient: c.rocksClient,
	}
}

// AddKeys добавляет ключи для удаления.
func (c *BatchDeleterRedis) AddKeys(keys ...string) {
	c.keys = append(c.keys, keys...)
}

// GetRocksCacheOptions возвращает конфигурационные опции по умолчанию для RocksCache.
func GetRocksCacheOptions() rockscache.Options {
	opts := rockscache.NewDefaultOptions()
	opts.StrongConsistency = true
	opts.RandomExpireAdjustment = 0.2

	return opts
}

// GetCache получает данные из кэша RocksCache или вызывает функцию для получения данных, если их нет в кэше.
func GetCache[T any](ctx context.Context, rcClient *rockscache.Client, key string, expire time.Duration, fn func(ctx context.Context) (T, error)) (T, error) {
	var t T
	var write bool
	// Пытаемся получить данные из кэша
	v, err := rcClient.Fetch2(ctx, key, expire, func() (s string, err error) {
		t, err = fn(ctx)
		if err != nil {
			return "", err
		}
		bs, err := json.Marshal(t)
		if err != nil {
			return "", err
		}
		write = true
		return string(bs), nil
	})
	if err != nil {
		return t, err
	}
	if write {
		return t, nil
	}
	if v == "" {
		return t, apperr.ErrDBRecordNotFound
	}
	err = json.Unmarshal([]byte(v), &t)
	if err != nil {
		return t, errors.New(fmt.Sprintf("cache json.Unmarshal failed, key:%s, value:%s, expire:%s %V", key, v, expire, err))
	}

	return t, nil
}

// BatchGetCache получает данные из кэша RocksCache для нескольких ключей или вызывает функцию для получения данных, если их нет в кэше.
func BatchGetCache[T any, K comparable](
	ctx context.Context,
	rcClient *rockscache.Client,
	expire time.Duration,
	keys []K,
	keyFn func(key K) string,
	fns func(ctx context.Context, key K) (T, error),
) ([]T, error) {
	if len(keys) == 0 {
		return nil, nil
	}
	res := make([]T, 0, len(keys))
	for _, key := range keys {
		val, err := GetCache(ctx, rcClient, keyFn(key), expire, func(ctx context.Context) (T, error) {
			return fns(ctx, key)
		})
		if err != nil {
			if apperr.Is(err, apperr.ErrDBRecordNotFound) {
				continue
			}
			return nil, err
		}
		res = append(res, val)
	}

	return res, nil
}
