package auth

import (
	"encoding/json"

	"github.com/RackSec/srslog"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin/binding"
	"github.com/rs/zerolog"
)

var (
	ErrSyslogUnavailable = apperr.New("syslog_unavailable", apperr.WithTextTranslate(translator.Translate{translator.RU: "Syslog %s недоступен: %s", translator.EN: "Syslog %s is unavailable: %s"}))
)

type Request struct {
	Level  int `json:"level" binding:"oneof=0 1 2 3 4 5"`
	Syslog []struct {
		Network string `json:"network" binding:"required,oneof=tcp udp"`
		Addr    string `json:"addr" binding:"required"`
	} `json:"syslog" binding:"required,dive"`
	File struct {
		Enabled    bool `json:"enabled"`
		MaxSizeMB  int  `json:"max_size_mb" binding:"required,min=1,max=65535"`
		MaxBackups int  `json:"max_backups" binding:"required,min=1,max=65535"`
		MaxAgeDays int  `json:"max_age_days" binding:"required,min=1,max=65535"`
		Compress   bool `json:"compress"`
	} `json:"file" binding:"required"`
	DB struct {
		Enabled bool `json:"enabled"`
	} `json:"db" binding:"required"`
}

type Transformer struct {
	Level  int `json:"level"`
	Syslog []struct {
		Network string `json:"network"`
		Addr    string `json:"addr"`
	} `json:"syslog"`
	File struct {
		Enabled    bool `json:"enabled"`
		MaxSizeMB  int  `json:"max_size_mb"`
		MaxBackups int  `json:"max_backups"`
		MaxAgeDays int  `json:"max_age_days"`
		Compress   bool `json:"compress"`
	} `json:"file"`
	DB struct {
		Enabled bool `json:"enabled"`
	} `json:"db" binding:"required"`
}

type Cfg struct {
	Level  int         `json:"level"`
	Syslog []CfgSyslog `json:"syslog"`
	File   CfgFile     `json:"file"`
	DB     CfgDB       `json:"db"`
}

type CfgSyslog struct {
	Network string `json:"network"`
	Addr    string `json:"addr"`
}

type CfgFile struct {
	Enabled    bool   `json:"enabled"`
	Path       string `json:"-"`
	MaxSizeMB  int    `json:"max_size_mb"`
	MaxBackups int    `json:"max_backups"`
	MaxAgeDays int    `json:"max_age_days"`
	Compress   bool   `json:"compress"`
}

type CfgDB struct {
	Enabled bool `json:"enabled"`
}

type Configurator struct {
	logPath string
	ch      chan Cfg
}

func NewConfigurator(logPath string) *Configurator {
	return &Configurator{
		logPath: logPath,
		ch:      make(chan Cfg, 1),
	}
}

func (c *Configurator) Watch() <-chan Cfg {
	return c.ch
}

func (c *Configurator) Action() string {
	return "Изменение настроек аудита"
}

func (c *Configurator) Unmarshal(data []byte) (any, error) {
	var cfg Cfg
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return Cfg{}, err
	}

	cfg.File.Path = c.logPath

	return cfg, nil
}

func (c *Configurator) Check(newData, lastData []byte) ([]byte, error) {
	var newCfg Request
	if err := binding.JSON.BindBody(newData, &newCfg); err != nil {
		return nil, apperr.ErrValidation.WithError(err)
	}

	cfg := Cfg{
		Level:  newCfg.Level,
		Syslog: []CfgSyslog{},
		File: CfgFile{
			Enabled:    newCfg.File.Enabled,
			MaxSizeMB:  newCfg.File.MaxSizeMB,
			MaxBackups: newCfg.File.MaxBackups,
			MaxAgeDays: newCfg.File.MaxAgeDays,
			Compress:   newCfg.File.Compress,
		},
		DB: CfgDB{
			Enabled: newCfg.DB.Enabled,
		},
	}

	for _, sys := range newCfg.Syslog {
		w, err := srslog.Dial(sys.Network, sys.Addr, srslog.LOG_LOCAL0|srslog.LOG_DEBUG, logger.AppName)
		if err != nil {
			return nil, ErrSyslogUnavailable.WithTextArgs(sys.Addr, err.Error()).WithError(err)
		}
		_ = w.Close()

		cfg.Syslog = append(cfg.Syslog, CfgSyslog{
			Network: sys.Network,
			Addr:    sys.Addr,
		})
	}

	return json.Marshal(cfg)
}

func (c *Configurator) AfterUpdate(data []byte) error {
	cfgU, err := c.Unmarshal(data)
	if err != nil {
		return err
	}

	cfg := cfgU.(Cfg)

	select {
	case c.ch <- cfg:
	default:
		select {
		case <-c.ch:
		default:
		}
		c.ch <- cfg
	}

	return nil
}

func (c *Configurator) Init() ([]byte, error) {
	cfg := Cfg{
		Level: int(zerolog.ErrorLevel),
		File: CfgFile{
			Enabled:    true,
			Path:       c.logPath,
			MaxSizeMB:  100,
			MaxBackups: 10,
			MaxAgeDays: 1095,
			Compress:   true,
		},
		Syslog: []CfgSyslog{},
		DB: CfgDB{
			Enabled: true,
		},
	}

	return json.Marshal(cfg)
}

func (c *Configurator) Transform(data []byte) ([]byte, error) {
	var cfg Transformer
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return json.Marshal(cfg)
}
