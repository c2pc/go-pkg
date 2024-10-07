package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

const (
	defaultBatchSize       = 50 // Default batch size for processing keys
	defaultConcurrentLimit = 3  // Default concurrency limit for processing
)

// ShardManager управляет шардированием и обработкой ключей
type ShardManager struct {
	redisClient redis.UniversalClient
	config      *RedisShardManagerConfig
}

// RedisShardManagerConfig конфигурация для ShardManager
type RedisShardManagerConfig struct {
	batchSize       int  // Размер пакета для обработки ключей
	continueOnError bool // Продолжать ли обработку при возникновении ошибок
	concurrentLimit int  // Лимит конкурентных операций
}

// Option это функция-конфигуратор для RedisShardManagerConfig
type Option func(c *RedisShardManagerConfig)

// NewRedisShardManager создает новый экземпляр ShardManager
func NewRedisShardManager(redisClient redis.UniversalClient, opts ...Option) *ShardManager {
	config := &RedisShardManagerConfig{
		batchSize:       defaultBatchSize,       // Установка размера пакета по умолчанию
		continueOnError: false,                  // Не продолжать обработку при ошибках по умолчанию
		concurrentLimit: defaultConcurrentLimit, // Лимит конкурентных операций по умолчанию
	}
	for _, opt := range opts {
		opt(config)
	}
	rsm := &ShardManager{
		redisClient: redisClient,
		config:      config,
	}
	return rsm
}

// WithBatchSize устанавливает количество ключей для обработки за раз
func WithBatchSize(size int) Option {
	return func(c *RedisShardManagerConfig) {
		c.batchSize = size
	}
}

// WithContinueOnError устанавливает, следует ли продолжать обработку при ошибках
func WithContinueOnError(continueOnError bool) Option {
	return func(c *RedisShardManagerConfig) {
		c.continueOnError = continueOnError
	}
}

// WithConcurrentLimit устанавливает лимит конкурентных операций
func WithConcurrentLimit(limit int) Option {
	return func(c *RedisShardManagerConfig) {
		c.concurrentLimit = limit
	}
}

// ProcessKeysBySlot группирует ключи по хэш-слотам Redis и обрабатывает их с использованием предоставленной функции.
func (rsm *ShardManager) ProcessKeysBySlot(
	ctx context.Context,
	keys []string,
	processFunc func(ctx context.Context, slot int64, keys []string) error,
) error {

	// Группировка ключей по слотам
	slots, err := groupKeysBySlot(ctx, rsm.redisClient, keys)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(rsm.config.concurrentLimit)

	// Обработка ключей в каждом слоте с использованием предоставленной функции
	for slot, singleSlotKeys := range slots {
		batches := splitIntoBatches(singleSlotKeys, rsm.config.batchSize)
		for _, batch := range batches {
			slot, batch := slot, batch // Избежание захвата переменных в замыкание
			g.Go(func() error {
				err := processFunc(ctx, slot, batch)
				if err != nil {
					if !rsm.config.continueOnError {
						return err
					}
				}
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

// groupKeysBySlot группирует ключи по хэш-слотам Redis.
func groupKeysBySlot(ctx context.Context, redisClient redis.UniversalClient, keys []string) (map[int64][]string, error) {
	slots := make(map[int64][]string)
	clusterClient, isCluster := redisClient.(*redis.ClusterClient)
	if isCluster {
		pipe := clusterClient.Pipeline()
		cmds := make([]*redis.IntCmd, len(keys))
		for i, key := range keys {
			cmds[i] = pipe.ClusterKeySlot(ctx, key)
		}
		_, err := pipe.Exec(ctx)
		if err != nil {
			return nil, err
		}

		for i, cmd := range cmds {
			slot, err := cmd.Result()
			if err != nil {
				return nil, err
			}
			slots[slot] = append(slots[slot], keys[i])
		}
	} else {
		// Если это не кластерный клиент, все ключи помещаются в один слот (0)
		slots[0] = keys
	}

	return slots, nil
}

// splitIntoBatches делит ключи на пакеты заданного размера
func splitIntoBatches(keys []string, batchSize int) [][]string {
	var batches [][]string
	for batchSize < len(keys) {
		keys, batches = keys[batchSize:], append(batches, keys[0:batchSize:batchSize])
	}
	return append(batches, keys)
}

// ProcessKeysBySlot группирует ключи по хэш-слотам Redis и обрабатывает их с использованием предоставленной функции.
func ProcessKeysBySlot(
	ctx context.Context,
	redisClient redis.UniversalClient,
	keys []string,
	processFunc func(ctx context.Context, slot int64, keys []string) error,
	opts ...Option,
) error {

	config := &RedisShardManagerConfig{
		batchSize:       defaultBatchSize,
		continueOnError: false,
		concurrentLimit: defaultConcurrentLimit,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Группировка ключей по слотам
	slots, err := groupKeysBySlot(ctx, redisClient, keys)
	if err != nil {
		return err
	}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(config.concurrentLimit)

	// Обработка ключей в каждом слоте с использованием предоставленной функции
	for slot, singleSlotKeys := range slots {
		batches := splitIntoBatches(singleSlotKeys, config.batchSize)
		for _, batch := range batches {
			slot, batch := slot, batch // Избежание захвата переменных в замыкание
			g.Go(func() error {
				err := processFunc(ctx, slot, batch)
				if err != nil {
					if !config.continueOnError {
						return err
					}
				}
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}
