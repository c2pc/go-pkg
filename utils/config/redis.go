package config

type Redis struct {
	Address     []string `yaml:"address"`      // Список адресов Redis серверов
	Username    string   `yaml:"username"`     // Имя пользователя для аутентификации Redis
	Password    string   `yaml:"password"`     // Пароль для аутентификации Redis
	ClusterMode bool     `yaml:"cluster_mode"` // Режим кластера Redis (true/false)
	DB          int      `yaml:"db"`           // Номер базы данных Redis (0 по умолчанию)
	MaxRetry    int      `yaml:"max_retry"`    // Максимальное количество попыток повторного подключения
}
