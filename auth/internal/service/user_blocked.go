package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/c2pc/go-pkg/v2/auth/internal/cache/cachekey"
	"github.com/c2pc/go-pkg/v2/auth/internal/fx"
	"github.com/c2pc/go-pkg/v2/auth/internal/i18n"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	"github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"github.com/c2pc/go-pkg/v2/utils/syslog"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	ErrUserBlockedNotFound = apperr.New("user_blocked_not_found", apperr.WithTextTranslate(i18n.ErrUserBlockedNotFound), apperr.WithCode(code.NotFound))
)

type IUserBlockedService interface {
	Trx(db *gorm.DB) IUserBlockedService
	List(ctx context.Context, m *meta.Meta[model.UserBlocked]) error
	Unlock(ctx context.Context, id int64) (string, error)
}

type UserBlockedService struct {
	repositories repository.Repositories
	cache        *fx.CacheHolder
}

func NewUserBlockedService(
	repositories repository.Repositories,
	cache *fx.CacheHolder,
) UserBlockedService {
	return UserBlockedService{
		repositories: repositories,
		cache:        cache,
	}
}

func (s UserBlockedService) Trx(db *gorm.DB) IUserBlockedService {
	s.repositories.UserBlockedRepository = s.repositories.UserBlockedRepository.Trx(db)
	return s
}

func (s UserBlockedService) List(ctx context.Context, m *meta.Meta[model.UserBlocked]) error {
	return s.repositories.UserBlockedRepository.With("User").Paginate(ctx, m, ``)
}

func (s UserBlockedService) Unlock(ctx context.Context, id int64) (userLogin string, err error) {
	defer func() {
		success := err == nil
		msg := fmt.Sprintf("Удаление временной блокировки администратора: %s", userLogin)
		if err != nil {
			msg = fmt.Sprintf("%s: %s", msg, err.Error())
		}

		syslog.Write(ctx, syslog.Record{EventID: "unlock-admin", EventName: "Удаление временной блокировки администратора", Severity: syslog.SeverityLow, Success: success}, msg)
	}()

	blocked, err := s.repositories.UserBlockedRepository.With("User").Find(ctx, `user_id = ?`, id)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return "", ErrUserBlockedNotFound
		}
		return "", err
	}

	if err := s.repositories.UserBlockedRepository.Delete(ctx, `user_id = ?`, blocked.UserID); err != nil {
		if !apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return blocked.User.Login, err
		}
	}

	err = s.cache.Get().LimiterCache.ResetAttempts(ctx, cachekey.GetUsernameKey()+blocked.User.Login)
	if err != nil && !errors.Is(err, redis.Nil) {
		return blocked.User.Login, apperr.ErrInternal.WithError(err)
	}

	return blocked.User.Login, nil
}
