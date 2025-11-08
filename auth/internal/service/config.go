package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/c2pc/go-pkg/v2/auth/configurator"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/syslog"
	"github.com/c2pc/go-pkg/v2/utils/translator"

	"gorm.io/gorm"
)

var (
	ErrConfigNotFound = apperr.New("config_not_found",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Конфигурация не найдена", translator.EN: "Config not found"}),
		apperr.WithCode(code.NotFound),
	)

	ErrConfigAlreadyExists = apperr.New("config_already_exists",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Конфигурация уже существует", translator.EN: "Config already exists"}),
		apperr.WithCode(code.AlreadyExists),
	)
)

type IConfigService interface {
	Trx(db *gorm.DB) IConfigService
	GetByKey(ctx context.Context, key string) (*model.Config, error)
	Update(ctx context.Context, key string, input json.RawMessage) (string, error)
	Init(ctx context.Context) error
	SetConfig(ctx context.Context, key string, cfg configurator.Configurator) (configurator.Config, error)
	GetWithoutTransform(ctx context.Context, key string) (*model.Config, error)
	UploadFiles(ctx context.Context, configKey string, input []UploadFilesInput) ([]model.ConfigFile, string, error)
}

type ConfigService struct {
	repositories repository.Repositories
	configs      configurator.Configs
}

func NewConfigService(repositories repository.Repositories) ConfigService {
	return ConfigService{
		repositories: repositories,
		configs:      make(configurator.Configs),
	}
}

func (s ConfigService) Trx(db *gorm.DB) IConfigService {
	s.repositories.ConfigRepository = s.repositories.ConfigRepository.Trx(db)
	s.repositories.ConfigFileRepository = s.repositories.ConfigFileRepository.Trx(db)
	return s
}

func (s ConfigService) SetConfig(ctx context.Context, key string, cfg configurator.Configurator) (configurator.Config, error) {
	cfgDB, err := s.repositories.ConfigRepository.With("Files").Find(ctx, `"key" = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			cfgDB, err = s.createDefault(ctx, key, cfg)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	s.configs[key] = cfg

	return cfgDB, nil
}

func (s ConfigService) Init(ctx context.Context) error {
	configsFromDB, err := s.repositories.ConfigRepository.List(ctx, &meta.Filter{}, ``)
	if err != nil {
		return err
	}

	for _, config := range configsFromDB {
		if _, ok := s.configs[config.Key]; !ok {
			err := s.delete(ctx, config.Key)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s ConfigService) createDefault(ctx context.Context, key string, cfg configurator.Configurator) (*model.Config, error) {
	jsonValue, err := cfg.Init()
	if err != nil {
		return nil, err
	}

	authConfig := &model.Config{
		Key:       key,
		Value:     jsonValue,
		UpdatedAt: time.Now().UTC(),
	}

	authConfig, err = s.repositories.ConfigRepository.Create(ctx, authConfig)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBDuplicated) {
			return nil, ErrConfigAlreadyExists
		}

		return nil, err
	}

	return authConfig, nil
}

func (s ConfigService) GetByKey(ctx context.Context, key string) (*model.Config, error) {
	authConfig, err := s.repositories.ConfigRepository.WithOne("Files", func(db *gorm.DB) *gorm.DB {
		return db.Omit("value")
	}).Find(ctx, `"key" = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrConfigNotFound
		}

		return nil, err
	}

	if cfg, ok := s.configs[authConfig.Key]; ok {
		authConfig.Value, err = cfg.Transform(authConfig)
		if err != nil {
			return nil, err
		}

		return authConfig, nil
	}

	return nil, ErrConfigNotFound
}

func (s ConfigService) GetWithoutTransform(ctx context.Context, key string) (*model.Config, error) {
	authConfig, err := s.repositories.ConfigRepository.With("Files").Find(ctx, `"key" = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrConfigNotFound
		}

		return nil, err
	}

	if _, ok := s.configs[authConfig.Key]; ok {
		return authConfig, nil
	}

	return nil, ErrConfigNotFound
}

func (s ConfigService) Update(ctx context.Context, key string, input json.RawMessage) (action string, err error) {
	defer func() {
		success := err == nil
		syslog.Write(ctx, syslog.Record{EventID: "update-config", EventName: "Изменение конфигурации системы", Severity: syslog.SeverityLow, Success: success}, action)
	}()

	cfg, ok := s.configs[key]
	if !ok {
		return "Изменение конфигурации системы", ErrConfigNotFound
	}

	cfgDB, err := s.repositories.ConfigRepository.With("Files").Find(ctx, `"key" = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return cfg.Action(), ErrConfigNotFound
		}

		return cfg.Action(), err
	}

	newCfg := &model.Config{
		Key:       key,
		Value:     input,
		UpdatedAt: time.Now().UTC(),
	}

	newCfg.Value, err = cfg.Check(newCfg, cfgDB)
	if err != nil {
		return cfg.Action(), err
	}

	err = s.repositories.ConfigRepository.Update(ctx, newCfg, []any{"value", "updated_at"}, `"key" = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return cfg.Action(), ErrConfigNotFound
		}

		return cfg.Action(), err
	}

	newCfg.Files = cfgDB.Files
	err = cfg.AfterUpdate(newCfg)
	if err != nil {
		return cfg.Action(), err
	}

	return cfg.Action(), nil
}

func (s ConfigService) delete(ctx context.Context, key string) error {
	return s.repositories.ConfigRepository.Delete(ctx, `"key" = ?`, key)
}

type UploadFilesInput struct {
	Key      string
	FileName string
	Data     []byte
}

func (s ConfigService) UploadFiles(ctx context.Context, configKey string, input []UploadFilesInput) ([]model.ConfigFile, string, error) {
	cfg, ok := s.configs[configKey]
	if !ok {
		return nil, "Изменение конфигурации системы", ErrConfigNotFound
	}

	cfgDB, err := s.repositories.ConfigRepository.Find(ctx, `"key" = ?`, configKey)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, cfg.Action(), ErrConfigNotFound
		}

		return nil, cfg.Action(), err
	}

	var files []model.ConfigFile
	for _, f := range input {
		files = append(files, model.ConfigFile{
			ConfigKey: cfgDB.Key,
			Key:       f.Key,
			Value:     f.Data,
			FileName:  f.FileName,
			UpdatedAt: time.Now().UTC(),
		})
	}

	for _, file := range files {
		_, err := s.repositories.ConfigFileRepository.
			CreateOrUpdate(ctx, &file, []interface{}{"key"}, []interface{}{"config_key", "value", "file_name", "updated_at"}, []interface{}{})
		if err != nil {
			return nil, cfg.Action(), err
		}
	}

	err = s.repositories.ConfigRepository.Update(ctx, &model.Config{UpdatedAt: time.Now().UTC()}, []any{"updated_at"}, `"key" = ?`, cfgDB.Key)
	if err != nil {
		return nil, cfg.Action(), err
	}

	return files, cfg.Action(), nil
}
