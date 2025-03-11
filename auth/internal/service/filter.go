package service

import (
	"context"

	"github.com/c2pc/go-pkg/v2/utils/mcontext"

	"github.com/c2pc/go-pkg/v2/auth/internal/i18n"
	"github.com/c2pc/go-pkg/v2/auth/internal/model"
	repository2 "github.com/c2pc/go-pkg/v2/auth/internal/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"gorm.io/gorm"
)

var (
	ErrFilterNotFound = apperr.New("filter_not_found", apperr.WithTextTranslate(i18n.ErrFilterNotFound), apperr.WithCode(code.NotFound))
	ErrFilterExists   = apperr.New("filter_exists_error", apperr.WithTextTranslate(i18n.ErrFilterExists), apperr.WithCode(code.InvalidArgument))
)

type IFilterService interface {
	Trx(db *gorm.DB) IFilterService
	List(ctx context.Context, m *model2.Meta[model.Filter]) error
	GetById(ctx context.Context, id int) (*model.Filter, error)
	Create(ctx context.Context, input FilterCreateInput) (*model.Filter, error)
	Update(ctx context.Context, id int, input FilterUpdateInput) error
	Delete(ctx context.Context, id int) error
}

type FilterService struct {
	filterRepository repository2.IFilterRepository
}

func NewFilterService(
	filterRepository repository2.IFilterRepository,
) FilterService {
	return FilterService{
		filterRepository: filterRepository,
	}
}

func (s FilterService) Trx(db *gorm.DB) IFilterService {
	s.filterRepository = s.filterRepository.Trx(db)
	return s
}

func (s FilterService) List(ctx context.Context, m *model2.Meta[model.Filter]) error {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	deviceID, ok := mcontext.GetOpDeviceID(ctx)
	if !ok {
		return apperr.ErrUnauthenticated.WithErrorText("operation device id is empty")
	}

	if err := s.filterRepository.Paginate(ctx, m, `user_id = ? AND device_id = ?`, userID, deviceID); err != nil {
		return err
	}

	return nil
}

func (s FilterService) GetById(ctx context.Context, id int) (*model.Filter, error) {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	deviceID, ok := mcontext.GetOpDeviceID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation device id is empty")
	}

	filter, err := s.filterRepository.Find(ctx, `id = ? AND user_id = ? AND device_id = ?`, id, userID, deviceID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrFilterNotFound
		}
		return nil, err
	}

	return filter, nil
}

type FilterCreateInput struct {
	Name     string
	Endpoint string
	Value    string
}

func (s FilterService) Create(ctx context.Context, input FilterCreateInput) (*model.Filter, error) {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	deviceID, ok := mcontext.GetOpDeviceID(ctx)
	if !ok {
		return nil, apperr.ErrUnauthenticated.WithErrorText("operation device id is empty")
	}

	filter, err := s.filterRepository.Create(ctx, &model.Filter{
		UserID:   userID,
		DeviceID: deviceID,
		Name:     input.Name,
		Endpoint: input.Endpoint,
		Value:    []byte(input.Value),
	}, "id")
	if err != nil {
		if apperr.Is(err, apperr.ErrDBDuplicated) {
			return nil, ErrFilterExists
		}
		return nil, err
	}

	return filter, nil
}

type FilterUpdateInput struct {
	Name  *string
	Value *string
}

func (s FilterService) Update(ctx context.Context, id int, input FilterUpdateInput) error {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	deviceID, ok := mcontext.GetOpDeviceID(ctx)
	if !ok {
		return apperr.ErrUnauthenticated.WithErrorText("operation device id is empty")
	}

	filter, err := s.filterRepository.Omit("value").Find(ctx, `id = ? AND user_id = ? AND device_id = ?`, id, userID, deviceID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrFilterNotFound
		}
		return err
	}

	var selects []interface{}
	if input.Name != nil && *input.Name != "" {
		filter.Name = *input.Name
		selects = append(selects, "name")
	}

	if input.Value != nil {
		filter.Value = []byte(*input.Value)
		selects = append(selects, "value")
	}

	if len(selects) > 0 {
		if err = s.filterRepository.Update(ctx, filter, selects, `id = ?`, filter.ID); err != nil {
			if apperr.Is(err, apperr.ErrDBDuplicated) {
				return ErrFilterExists
			}
			return err
		}
	}

	return nil
}

func (s FilterService) Delete(ctx context.Context, id int) error {
	userID, ok := mcontext.GetOpUserID(ctx)
	if !ok {
		return apperr.ErrUnauthenticated.WithErrorText("operation user id is empty")
	}

	deviceID, ok := mcontext.GetOpDeviceID(ctx)
	if !ok {
		return apperr.ErrUnauthenticated.WithErrorText("operation device id is empty")
	}

	filter, err := s.filterRepository.Omit("value").Find(ctx, `id = ? AND user_id = ? AND device_id = ?`, id, userID, deviceID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrFilterNotFound
		}
		return err
	}

	if err := s.filterRepository.Delete(ctx, `id = ?`, filter.ID); err != nil {
		return err
	}

	return nil
}
