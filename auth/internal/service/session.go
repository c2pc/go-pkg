package service

import (
	"context"
	"errors"

	"github.com/c2pc/go-pkg/v2/auth/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/i18n"
	model2 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/constant"
	model3 "github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound = apperr.New("session_not_found", apperr.WithTextTranslate(i18n.ErrSessionNotFound), apperr.WithCode(code.NotFound))
)

type ISessionService interface {
	Trx(db *gorm.DB) ISessionService
	List(ctx context.Context, m *model3.Meta[model2.RefreshToken]) error
	End(ctx context.Context, id int) (string, error)
}

type SessionService struct {
	repositories repository.Repositories
	cache        *fx.CacheHolder
}

func NewSessionService(
	repositories repository.Repositories,
	cache *fx.CacheHolder,
) SessionService {
	return SessionService{
		repositories: repositories,
		cache:        cache,
	}
}

func (s SessionService) Trx(db *gorm.DB) ISessionService {
	s.repositories.TokenRepository = s.repositories.TokenRepository.Trx(db)
	return s
}

func (s SessionService) List(ctx context.Context, m *model3.Meta[model2.RefreshToken]) error {
	return s.repositories.TokenRepository.With("user").Paginate(ctx, m, ``)
}

func (s SessionService) End(ctx context.Context, id int) (string, error) {
	token, err := s.repositories.TokenRepository.With("User").FindById(ctx, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return "", ErrSessionNotFound
		}
		return "", err
	}

	if err := s.repositories.TokenRepository.Delete(ctx, `id = ?`, id); err != nil {
		if !apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return token.User.Login, err
		}
	}

	m, err := s.cache.Get().TokenCache.GetTokensWithoutError(ctx, token.UserID, token.DeviceID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return token.User.Login, apperr.ErrInternal.WithError(err)
	}
	for k := range m {
		m[k] = constant.KickedToken
		err = s.cache.Get().TokenCache.SetTokenMapByUidPid(ctx, token.UserID, token.DeviceID, m)
		if err != nil {
			return token.User.Login, apperr.ErrInternal.WithError(err)
		}
	}

	if err := s.cache.Get().UserCache.DelUsersInfo(token.UserID).ChainExecDel(ctx); err != nil {
		return token.User.Login, apperr.ErrInternal.WithError(err)
	}

	return token.User.Login, nil
}
