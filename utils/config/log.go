package config

type LOG struct {
	Debug      string `yaml:"debug,omitempty"`
	Folder     string `yaml:"folder"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}
