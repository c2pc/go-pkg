package service

import (
	"context"
	"encoding/json"

	"github.com/c2pc/go-pkg/v2/auth_config/configurator"
	"github.com/c2pc/go-pkg/v2/auth_config/internal/model"
	"github.com/c2pc/go-pkg/v2/auth_config/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/translator"

	"gorm.io/gorm"
)

var (
	ErrAuthConfigNotFound = apperr.New("auth_config_not_found",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Конфигурация не найдена", translator.EN: "Config not found"}),
		apperr.WithCode(code.NotFound),
	)

	ErrAuthConfigAlreadyExists = apperr.New("auth_config_already_exists",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Конфигурация уже существует", translator.EN: "Config already exists"}),
		apperr.WithCode(code.AlreadyExists),
	)
)

type IAuthConfigService interface {
	Trx(db *gorm.DB) IAuthConfigService
	List(ctx context.Context) ([]model.AuthConfig, error)
	GetByKey(ctx context.Context, key string) (*model.AuthConfig, error)
	Update(ctx context.Context, key string, input json.RawMessage) (string, error)
	Init(ctx context.Context) error
	SetConfig(ctx context.Context, key string, cfg configurator.Configurator) error
	GetWithoutTransform(ctx context.Context, key string) (*model.AuthConfig, error)
}

type AuthConfigService struct {
	authConfigRepo repository.AuthConfigRepository
	configs        configurator.Configs
}

func NewAuthConfigService(repo repository.AuthConfigRepository) AuthConfigService {
	return AuthConfigService{
		authConfigRepo: repo,
		configs:        make(configurator.Configs),
	}
}

func (s AuthConfigService) Trx(db *gorm.DB) IAuthConfigService {
	s.authConfigRepo = s.authConfigRepo.Trx(db)
	return s
}

func (s AuthConfigService) SetConfig(ctx context.Context, key string, cfg configurator.Configurator) error {
	_, err := s.authConfigRepo.Find(ctx, `key = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			_, err = s.createDefault(ctx, key, cfg)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	s.configs[key] = cfg

	return nil
}

func (s AuthConfigService) Init(ctx context.Context) error {
	configsFromDB, err := s.authConfigRepo.List(ctx, &model2.Filter{}, ``)
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

func (s AuthConfigService) createDefault(ctx context.Context, key string, cfg configurator.Configurator) (*model.AuthConfig, error) {
	jsonValue, err := cfg.Init()
	if err != nil {
		return nil, err
	}

	authConfig := &model.AuthConfig{
		Key:   key,
		Value: jsonValue,
	}

	authConfig, err = s.authConfigRepo.Create(ctx, authConfig)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBDuplicated) {
			return nil, ErrAuthConfigAlreadyExists
		}

		return nil, err
	}

	return authConfig, nil
}

func (s AuthConfigService) List(ctx context.Context) ([]model.AuthConfig, error) {
	data, err := s.authConfigRepo.List(ctx, &model2.Filter{}, ``)
	if err != nil {
		return nil, err
	}

	rows := make([]model.AuthConfig, 0)
	for _, row := range data {
		if cfg, ok := s.configs[row.Key]; ok {
			v, err := cfg.Transform(row.Value)
			if err != nil {
				return nil, err
			}

			rows = append(rows, model.AuthConfig{
				Key:       row.Key,
				Value:     v,
				UpdatedAt: row.UpdatedAt,
			})
		}
	}

	return rows, nil
}

func (s AuthConfigService) GetByKey(ctx context.Context, key string) (*model.AuthConfig, error) {
	authConfig, err := s.authConfigRepo.Find(ctx, `key = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrAuthConfigNotFound
		}

		return nil, err
	}

	if cfg, ok := s.configs[authConfig.Key]; ok {
		v, err := cfg.Transform(authConfig.Value)
		if err != nil {
			return nil, err
		}

		return &model.AuthConfig{Key: key, Value: v, UpdatedAt: authConfig.UpdatedAt}, nil
	}

	return nil, ErrAuthConfigNotFound
}

func (s AuthConfigService) GetWithoutTransform(ctx context.Context, key string) (*model.AuthConfig, error) {
	authConfig, err := s.authConfigRepo.Find(ctx, `key = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrAuthConfigNotFound
		}

		return nil, err
	}

	if _, ok := s.configs[authConfig.Key]; ok {
		return &model.AuthConfig{Key: key, Value: authConfig.Value, UpdatedAt: authConfig.UpdatedAt}, nil
	}

	return nil, ErrAuthConfigNotFound
}

func (s AuthConfigService) Update(ctx context.Context, key string, input json.RawMessage) (string, error) {
	cfg, ok := s.configs[key]
	if !ok {
		return "Изменение конфигурации системы", ErrAuthConfigNotFound
	}

	cfgDB, err := s.authConfigRepo.Find(ctx, `key = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return cfg.Action(), ErrAuthConfigNotFound
		}

		return cfg.Action(), err
	}

	jsonValue, err := cfg.Check(input, cfgDB.Value)
	if err != nil {
		return cfg.Action(), err
	}

	err = s.authConfigRepo.Update(ctx, &model.AuthConfig{Key: key, Value: jsonValue}, []any{"value"}, `key = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return cfg.Action(), ErrAuthConfigNotFound
		}

		return cfg.Action(), err
	}

	err = cfg.AfterUpdate(jsonValue)
	if err != nil {
		return cfg.Action(), err
	}

	return cfg.Action(), nil
}

func (s AuthConfigService) delete(ctx context.Context, key string) error {
	return s.authConfigRepo.Delete(ctx, `key = ?`, key)
}
