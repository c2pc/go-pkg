package config

type HTTP struct {
	Host string `yaml:"host,omitempty"` // Порт может быть пустым и будет опущен из сериализованного YAML, если он не установлен
	Port string `yaml:"port"`           // Порт сервера (например, "8080")
}
