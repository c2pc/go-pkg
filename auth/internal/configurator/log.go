package configurator

import (
	"encoding/json"

	"github.com/c2pc/go-pkg/v2/auth/configurator"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/logger"
	"github.com/gin-gonic/gin/binding"
)

type LogRequest struct {
	Level      *int8 `json:"level" binding:"omitempty,min=0,max=4"`
	MaxSizeMB  *int  `json:"max_size_mb" binding:"omitempty,min=1,max=65535"`
	MaxBackups *int  `json:"max_backups" binding:"omitempty,min=1,max=65535"`
	MaxAgeDays *int  `json:"max_age_days" binding:"omitempty,min=1,max=65535"`
	Compress   *bool `json:"compress"  binding:"omitempty"`
}

type LogTransformer struct {
	Level      int8 `json:"level"`
	MaxSizeMB  int  `json:"max_size_mb"`
	MaxBackups int  `json:"max_backups"`
	MaxAgeDays int  `json:"max_age_days"`
	Compress   bool `json:"compress"`
}

type LogCfg struct {
	Level      int8   `json:"level"`
	Path       string `json:"-"`
	MaxSizeMB  int    `json:"max_size_mb"`
	MaxBackups int    `json:"max_backups"`
	MaxAgeDays int    `json:"max_age_days"`
	Compress   bool   `json:"compress"`
}

type LogConfigurator struct {
	logPath string
	ch      chan LogCfg
}

func NewLogConfigurator(logPath string) *LogConfigurator {
	return &LogConfigurator{
		logPath: logPath,
		ch:      make(chan LogCfg, 1),
	}
}

func (c *LogConfigurator) Watch() <-chan LogCfg {
	return c.ch
}

func (c *LogConfigurator) Action() string {
	return "Изменение настроек логирования"
}

func (c *LogConfigurator) Unmarshal(data configurator.Config) (any, error) {
	var cfg LogCfg
	err := json.Unmarshal(data.GetValue(), &cfg)
	if err != nil {
		return LogCfg{}, err
	}

	cfg.Path = c.logPath

	return cfg, nil
}

func (c *LogConfigurator) Check(newData, lastData configurator.Config) ([]byte, error) {
	var newCfg LogRequest
	if err := binding.JSON.BindBody(newData.GetValue(), &newCfg); err != nil {
		return nil, apperr.ErrValidation.WithError(err)
	}

	var lastCfg LogCfg
	err := json.Unmarshal(lastData.GetValue(), &lastCfg)
	if err != nil {
		return nil, err
	}

	if newCfg.Level != nil {
		lastCfg.Level = *newCfg.Level
	}
	if newCfg.MaxSizeMB != nil {
		lastCfg.MaxSizeMB = *newCfg.MaxSizeMB
	}
	if newCfg.MaxBackups != nil {
		lastCfg.MaxBackups = *newCfg.MaxBackups
	}
	if newCfg.MaxAgeDays != nil {
		lastCfg.MaxAgeDays = *newCfg.MaxAgeDays
	}
	if newCfg.Compress != nil {
		lastCfg.Compress = *newCfg.Compress
	}

	return json.Marshal(lastCfg)
}

func (c *LogConfigurator) AfterUpdate(data configurator.Config) error {
	cfgU, err := c.Unmarshal(data)
	if err != nil {
		return err
	}

	cfg := cfgU.(LogCfg)

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

func (c *LogConfigurator) Init() ([]byte, error) {
	cfg := LogCfg{
		Level:      int8(logger.DebugLevel),
		Path:       c.logPath,
		MaxSizeMB:  20,
		MaxBackups: 10,
		MaxAgeDays: 30,
		Compress:   true,
	}

	return json.Marshal(cfg)
}

func (c *LogConfigurator) Transform(data configurator.Config) ([]byte, error) {
	var cfg LogTransformer
	err := json.Unmarshal(data.GetValue(), &cfg)
	if err != nil {
		return nil, err
	}

	return json.Marshal(cfg)
}
