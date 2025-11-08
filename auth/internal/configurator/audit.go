package configurator

import (
	"encoding/json"

	"github.com/RackSec/srslog"
	"github.com/c2pc/go-pkg/v2/auth/configurator"
	"github.com/c2pc/go-pkg/v2/utils/app_data"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/gin-gonic/gin/binding"
)

var (
	ErrSyslogUnavailable = apperr.New("syslog_unavailable", apperr.WithTextTranslate(translator.Translate{translator.RU: "Syslog %s недоступен: %s", translator.EN: "Syslog %s is unavailable: %s"}))
)

type AuditRequest struct {
	File *struct {
		MaxSizeMB  int  `json:"max_size_mb" binding:"required,min=1,max=65535"`
		MaxBackups int  `json:"max_backups" binding:"required,min=1,max=65535"`
		MaxAgeDays int  `json:"max_age_days" binding:"required,min=1,max=65535"`
		Compress   bool `json:"compress"`
		Syslog     []struct {
			Network string `json:"network" binding:"required,oneof=tcp udp"`
			Addr    string `json:"addr" binding:"required"`
		} `json:"syslog" binding:"required,dive"`
	} `json:"file" binding:"omitempty"`
	DB *struct {
		Admin struct {
			MaxAgeDays         int     `json:"max_age_days" binding:"required,min=1,max=65535"`
			MaxCount           int     `json:"max_count" binding:"required,min=1,max=65535"`
			PercentageOfDelete float64 `json:"percentage_of_delete" binding:"required,gte=1,lte=100"`
		} `json:"admin" binding:"omitempty"`
	} `json:"db" binding:"omitempty"`
}

type AuditTransformer struct {
	File struct {
		MaxSizeMB  int  `json:"max_size_mb"`
		MaxBackups int  `json:"max_backups"`
		MaxAgeDays int  `json:"max_age_days"`
		Compress   bool `json:"compress"`
		Syslog     []struct {
			Network string `json:"network"`
			Addr    string `json:"addr"`
		} `json:"syslog"`
	} `json:"file"`
	DB struct {
		Admin struct {
			MaxAgeDays         int     `json:"max_age_days"`
			MaxCount           int     `json:"max_count"`
			PercentageOfDelete float64 `json:"percentage_of_delete"`
		} `json:"admin"`
	} `json:"db"`
}

type AuditCfg struct {
	File AuditCfgFile `json:"file"`
	DB   AuditCfgDB   `json:"db"`
}

type AuditCfgSyslog struct {
	Network string `json:"network"`
	Addr    string `json:"addr"`
}

type AuditCfgFile struct {
	Path       string           `json:"-"`
	MaxSizeMB  int              `json:"max_size_mb"`
	MaxBackups int              `json:"max_backups"`
	MaxAgeDays int              `json:"max_age_days"`
	Compress   bool             `json:"compress"`
	Syslog     []AuditCfgSyslog `json:"syslog"`
}

type AuditCfgDB struct {
	Admin AuditCfgDBSettings `json:"admin"`
}

type AuditCfgDBSettings struct {
	MaxAgeDays         int     `json:"max_age_days"`
	MaxCount           int     `json:"max_count"`
	PercentageOfDelete float64 `json:"percentage_of_delete"`
	LogDisabled        bool    `json:"logger_disabled"`
}

type AuditConfigurator struct {
	logPath string
	ch      chan AuditCfg
}

func NewAuditConfigurator(logPath string) *AuditConfigurator {
	return &AuditConfigurator{
		logPath: logPath,
		ch:      make(chan AuditCfg, 1),
	}
}

func (c *AuditConfigurator) Watch() <-chan AuditCfg {
	return c.ch
}

func (c *AuditConfigurator) Action() string {
	return "Изменение настроек аудита"
}

func (c *AuditConfigurator) Unmarshal(data configurator.Config) (any, error) {
	var cfg AuditCfg
	err := json.Unmarshal(data.GetValue(), &cfg)
	if err != nil {
		return AuditCfg{}, err
	}

	cfg.File.Path = c.logPath

	return cfg, nil
}

func (c *AuditConfigurator) Check(newData, lastData configurator.Config) ([]byte, error) {
	var newCfg AuditRequest
	if err := binding.JSON.BindBody(newData.GetValue(), &newCfg); err != nil {
		return nil, apperr.ErrValidation.WithError(err)
	}

	var lastCfg AuditCfg
	err := json.Unmarshal(lastData.GetValue(), &lastCfg)
	if err != nil {
		return nil, err
	}

	if newCfg.File != nil {
		lastCfg.File = AuditCfgFile{
			MaxSizeMB:  newCfg.File.MaxSizeMB,
			MaxBackups: newCfg.File.MaxBackups,
			MaxAgeDays: newCfg.File.MaxAgeDays,
			Compress:   newCfg.File.Compress,
			Syslog:     []AuditCfgSyslog{},
		}
		for _, sys := range newCfg.File.Syslog {
			w, err := srslog.Dial(sys.Network, sys.Addr, srslog.LOG_LOCAL0|srslog.LOG_DEBUG, app_data.AppName)
			if err != nil {
				return nil, ErrSyslogUnavailable.WithTextArgs(sys.Addr, err.Error()).WithError(err)
			}
			_ = w.Close()

			lastCfg.File.Syslog = append(lastCfg.File.Syslog, AuditCfgSyslog{
				Network: sys.Network,
				Addr:    sys.Addr,
			})
		}
	}

	if newCfg.DB != nil {
		lastCfg.DB = AuditCfgDB{
			Admin: AuditCfgDBSettings{
				MaxAgeDays:         newCfg.DB.Admin.MaxAgeDays,
				MaxCount:           newCfg.DB.Admin.MaxCount,
				PercentageOfDelete: newCfg.DB.Admin.PercentageOfDelete,
			},
		}
	}

	return json.Marshal(lastCfg)
}

func (c *AuditConfigurator) AfterUpdate(data configurator.Config) error {
	cfgU, err := c.Unmarshal(data)
	if err != nil {
		return err
	}

	cfg := cfgU.(AuditCfg)

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

func (c *AuditConfigurator) Init() ([]byte, error) {
	cfg := AuditCfg{
		File: AuditCfgFile{
			Path:       c.logPath,
			MaxSizeMB:  20,
			MaxBackups: 10,
			MaxAgeDays: 1095,
			Compress:   true,
			Syslog:     []AuditCfgSyslog{},
		},
		DB: AuditCfgDB{
			Admin: AuditCfgDBSettings{
				MaxAgeDays:         1095,
				MaxCount:           100000,
				PercentageOfDelete: 20,
			},
		},
	}

	return json.Marshal(cfg)
}

func (c *AuditConfigurator) Transform(data configurator.Config) ([]byte, error) {
	var cfg AuditTransformer
	err := json.Unmarshal(data.GetValue(), &cfg)
	if err != nil {
		return nil, err
	}

	return json.Marshal(cfg)
}
