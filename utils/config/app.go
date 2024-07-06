package config

type APP struct {
	Debug  string `yaml:"debug,omitempty"`
	LogDir string `yaml:"log_dir"`
}
