package config

// APP содержит конфигурацию приложения.
type APP struct {
	Debug  string `yaml:"debug,omitempty"` // Опциональный режим отладки
	LogDir string `yaml:"log_dir"`         // Директория для логов
}
