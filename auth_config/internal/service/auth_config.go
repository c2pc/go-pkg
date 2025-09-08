package service

import (
	"context"
	"encoding/json"

	"github.com/c2pc/go-pkg/v2/auth_config/internal/model"
	"github.com/c2pc/go-pkg/v2/auth_config/internal/repository"
	"github.com/c2pc/go-pkg/v2/auth_config/transformer"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/c2pc/go-pkg/v2/utils/translator"

	"gorm.io/gorm"
)

var (
	ErrAuthConfigNotFound = apperr.New("auth_config_not_found",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Конфигурация не найдена", translator.EN: "Auth config not found"}),
		apperr.WithCode(code.NotFound),
	)

	ErrAuthConfigAlreadyExists = apperr.New("auth_config_already_exists",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Конфигурация уже существует", translator.EN: "Auth config already exists"}),
		apperr.WithCode(code.AlreadyExists),
	)

	ErrKeyNotFound = apperr.New("key_not_found",
		apperr.WithTextTranslate(translator.Translate{translator.RU: "Ключ не найден", translator.EN: "Key not found"}),
		apperr.WithCode(code.NotFound),
	)
)

type IAuthConfigService interface {
	Trx(db *gorm.DB) IAuthConfigService
	CreateDefault(ctx context.Context, key string) (*model.AuthConfig, error)
	List(ctx context.Context) ([]model.AuthConfig, error)
	GetByKey(ctx context.Context, key string) (*model.AuthConfig, error)
	Update(ctx context.Context, key string, input json.RawMessage) error
}

type AuthConfigService struct {
	authConfigRepo         repository.AuthConfigRepository
	authConfigTransformers transformer.AuthConfigTransformers
}

func NewAuthConfigService(repo repository.AuthConfigRepository, tmpls transformer.AuthConfigTransformers) (IAuthConfigService, error) {
	service := AuthConfigService{
		authConfigRepo:         repo,
		authConfigTransformers: tmpls,
	}

	err := service.Init()
	if err != nil {
		return nil, err
	}

	return service, nil
}

func (s *AuthConfigService) Init() error {
	configsFromDB, err := s.authConfigRepo.List(context.Background(), &model2.Filter{}, ``)
	if err != nil {
		return err
	}

	for _, config := range configsFromDB {
		if _, ok := s.authConfigTransformers[config.Key]; !ok {
			err := s.Delete(context.Background(), config.Key)
			if err != nil {
				return err
			}
		}
	}

	for key := range s.authConfigTransformers {
		_, err := s.GetByKey(context.Background(), key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s AuthConfigService) Trx(db *gorm.DB) IAuthConfigService {
	s.authConfigRepo = s.authConfigRepo.Trx(db)
	return s
}

func (s AuthConfigService) CreateDefault(ctx context.Context, key string) (*model.AuthConfig, error) {
	trans := s.authConfigTransformers[key]
	if trans == nil {
		return nil, ErrAuthConfigNotFound
	}

	jsonValue, err := trans.Init()
	if err != nil {
		return nil, err
	}

	authConfig, err := s.authConfigRepo.Create(ctx, &model.AuthConfig{
		Key:   key,
		Value: json.RawMessage(jsonValue),
	})
	if err != nil {
		if apperr.Is(err, apperr.ErrDBDuplicated) {
			return nil, ErrAuthConfigAlreadyExists
		}

		return nil, err
	}

	return authConfig, nil
}

func (s AuthConfigService) List(ctx context.Context) ([]model.AuthConfig, error) {
	return s.authConfigRepo.List(ctx, &model2.Filter{}, ``)
}

func (s AuthConfigService) GetByKey(ctx context.Context, key string) (*model.AuthConfig, error) {
	authConfig, err := s.authConfigRepo.Find(ctx, `key = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return s.CreateDefault(ctx, key)
		}

		return nil, err
	}

	return authConfig, nil
}

func (s AuthConfigService) Update(ctx context.Context, key string, input json.RawMessage) error {
	trans := s.authConfigTransformers[key]
	if trans == nil {
		return ErrKeyNotFound
	}

	err := trans.Check(input)
	if err != nil {
		return err
	}

	_, err = s.GetByKey(ctx, key)
	if err != nil {
		return err
	}

	err = s.authConfigRepo.Update(ctx, &model.AuthConfig{Key: key, Value: input}, []any{"value"}, `key = ?`, key)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrAuthConfigNotFound
		}

		return err
	}

	err = trans.AfterUpdate(input)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthConfigService) Delete(ctx context.Context, key string) error {
	return s.authConfigRepo.Delete(ctx, `key = ?`, key)
}
