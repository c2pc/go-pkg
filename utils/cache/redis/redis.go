package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/c2pc/go-pkg/v2/utils/level"
	"github.com/redis/go-redis/v9"
)

// RedisClient представляет конфигурацию для клиента Redis, включая
// параметры для соединений в режиме одиночного узла и кластера.
type RedisClient struct {
	ClusterMode bool     // Флаг, указывающий, использовать ли Redis в режиме кластера.
	Address     []string // Список адресов серверов Redis (хост:порт).
	Username    string   // Имя пользователя для аутентификации в Redis (ACL Redis 6).
	Password    string   // Пароль для аутентификации в Redis.
	MaxRetry    int      // Максимальное количество попыток повторного выполнения команды.
	DB          int      // Номер базы данных для подключения (для режима одиночного узла).
	PoolSize    int      // Количество подключений в пуле.
}

// NewRedisClient создает новый экземпляр клиента Redis в зависимости от конфигурации.
func NewRedisClient(config *RedisClient, debug string) (redis.UniversalClient, error) {
	// Проверка наличия адресов Redis
	if len(config.Address) == 0 {
		return nil, errors.New("redis address is empty")
	}

	var cli redis.UniversalClient

	// Выбор конфигурации для клиента в зависимости от режима кластера
	if config.ClusterMode || len(config.Address) > 1 {
		// Конфигурация для клиента Redis-кластера
		opt := &redis.ClusterOptions{
			Addrs:      config.Address,
			Username:   config.Username,
			Password:   config.Password,
			PoolSize:   config.PoolSize,
			MaxRetries: config.MaxRetry,
		}
		cli = redis.NewClusterClient(opt)
	} else {
		// Конфигурация для одиночного клиента Redis
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

	// Установка логгера для клиента Redis, если режим отладки активен
	if level.Is(debug, level.TEST) {
		redis.SetLogger(defaultLogger())
	}

	// Проверка подключения к Redis
	if err := cli.Ping(context.Background()).Err(); err != nil {
		return nil, errors.New(fmt.Sprintf("Redis Ping failed %v", err))
	}

	return cli, nil
}
