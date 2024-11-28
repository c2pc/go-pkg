package service

import (
	"context"
	"errors"
	"time"

	cache2 "github.com/c2pc/go-pkg/v2/auth/internal/cache"
	"github.com/c2pc/go-pkg/v2/auth/internal/i18n"
	model2 "github.com/c2pc/go-pkg/v2/auth/internal/model"
	repository2 "github.com/c2pc/go-pkg/v2/auth/internal/repository"
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
	End(ctx context.Context, id int) error
}

type SessionService struct {
	tokenRepository repository2.ITokenRepository
	tokenCache      cache2.ITokenCache
	userCache       cache2.IUserCache
	refreshExpire   time.Duration
	db              *gorm.DB
}

func NewSessionService(
	tokenRepository repository2.ITokenRepository,
	tokenCache cache2.ITokenCache,
	userCache cache2.IUserCache,
	refreshExpire time.Duration,
) SessionService {
	return SessionService{
		tokenRepository: tokenRepository,
		tokenCache:      tokenCache,
		userCache:       userCache,
		refreshExpire:   refreshExpire,
	}
}

func (s SessionService) Trx(db *gorm.DB) ISessionService {
	s.tokenRepository = s.tokenRepository.Trx(db)
	s.db = db
	return s
}

func (s SessionService) List(ctx context.Context, m *model3.Meta[model2.RefreshToken]) error {
	return s.tokenRepository.With("user").Paginate(ctx, m, ``)
}

func (s SessionService) End(ctx context.Context, id int) error {
	token, err := s.tokenRepository.FindById(ctx, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	if err := s.tokenRepository.Delete(ctx, `id = ?`, id); err != nil {
		if !apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return err
		}
	}

	m, err := s.tokenCache.GetTokensWithoutError(ctx, token.UserID, token.DeviceID)
	if err != nil && !errors.Is(err, redis.Nil) {
		return apperr.ErrInternal.WithError(err)
	}
	for k := range m {
		m[k] = constant.KickedToken
		err = s.tokenCache.SetTokenMapByUidPid(ctx, token.UserID, token.DeviceID, m)
		if err != nil {
			return apperr.ErrInternal.WithError(err)
		}
	}

	if err := s.userCache.DelUsersInfo(token.UserID).ChainExecDel(ctx); err != nil {
		return apperr.ErrInternal.WithError(err)
	}

	return nil
}
