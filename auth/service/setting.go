package service

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/model"
	"github.com/c2pc/go-pkg/v2/auth/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/mcontext"
	"gorm.io/gorm"
)

type ISettingService interface {
	Trx(db *gorm.DB) ISettingService
	Get(ctx context.Context) (*model.Setting, error)
	Update(ctx context.Context, input SettingUpdateInput) error
}

type SettingService struct {
	settingRepository repository.ISettingRepository
}

func NewSettingService(
	settingRepository repository.ISettingRepository,
) SettingService {
	return SettingService{
		settingRepository: settingRepository,
	}
}

func (s SettingService) Trx(db *gorm.DB) ISettingService {
	s.settingRepository = s.settingRepository.Trx(db)
	return s
}

func (s SettingService) Get(ctx context.Context) (*model.Setting, error) {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	deviceID, ok := mcontext.GetOpDeviceID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation device id is empty")
	}

	setting, err := s.settingRepository.FirstOrCreate(ctx, &model.Setting{
		UserID:   userID,
		DeviceID: deviceID,
		Settings: nil,
	}, "", `user_id = ? AND device_id = ?`, userID, deviceID)
	if err != nil {
		return nil, err
	}

	return setting, nil
}

type SettingUpdateInput struct {
	Settings *string
}

func (s SettingService) Update(ctx context.Context, input SettingUpdateInput) error {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	deviceID, ok := mcontext.GetOpDeviceID(ctx)
	if !ok {
		return apperr.ErrUnauthenticated.WithErrorText("operation device id is empty")
	}

	setting := &model.Setting{
		UserID:   userID,
		DeviceID: deviceID,
		Settings: nil,
	}

	var selects []interface{}
	if input.Settings != nil {
		if *input.Settings == "" {
			setting.Settings = nil
		} else {
			setting.Settings = input.Settings
		}
		selects = append(selects, "settings")
	}

	if len(selects) > 0 {
		if err := s.settingRepository.Update(ctx, setting, selects, `user_id = ? AND device_id = ?`, userID, deviceID); err != nil {
			return err
		}
	}

	return nil
}
