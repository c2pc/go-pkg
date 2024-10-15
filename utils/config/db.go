package config

type DB struct {
	DSN         string `yaml:"dsn,omitempty"`
	MaxIdleConn int    `yaml:"max_idle_conn,omitempty"`
	MaxOpenConn int    `yaml:"max_open_conn,omitempty"`
}
