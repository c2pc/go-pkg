package config

type LOG struct {
	Debug      string `yaml:"debug,omitempty"`
	Directory  string `yaml:"directory"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}
