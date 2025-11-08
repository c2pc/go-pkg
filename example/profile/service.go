package profile

import (
	"context"

	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/meta"
	"gorm.io/gorm"
)

type Service struct {
	profileRepository IRepository
}

func NewService(
	profileRepository IRepository,
) Service {
	return Service{
		profileRepository: profileRepository,
	}
}

func (s Service) Trx(db *gorm.DB) profile.IProfileService {
	s.profileRepository = s.profileRepository.Trx(db)

	return s
}

func (s Service) GetById(ctx context.Context, userID int64) (profile.IModel, error) {
	prof, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, profile.ErrNotFound
		}
		return nil, err
	}

	return prof, nil
}

func (s Service) GetByIds(ctx context.Context, userID ...int64) ([]profile.IModel, error) {
	profs, err := s.profileRepository.List(ctx, &meta.Filter{}, `user_id IN (?)`, userID)
	if err != nil {
		return nil, err
	}

	m := make([]profile.IModel, len(profs))
	for i, prof := range profs {
		m[i] = prof
	}

	return m, nil
}

type CreateInput struct {
	Age     *int
	Height  *int
	Address string
}

func (s Service) Create(ctx context.Context, userID int64, input any) (profile.IModel, error) {
	inp := input.(*CreateInput)

	prof, err := s.profileRepository.Create(ctx, &Profile{
		Age:     inp.Age,
		Height:  inp.Height,
		Address: inp.Address,
		UserID:  userID,
	}, "id")
	if err != nil {
		if apperr.Is(err, apperr.ErrDBDuplicated) {
			return nil, profile.ErrExists
		}
		return nil, err
	}

	prof, err = s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		return nil, err
	}

	return prof, nil
}

type UpdateInput struct {
	Age     *int
	Height  *int
	Address *string
}

func (s Service) Update(ctx context.Context, userID int64, input any) error {
	prof, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return profile.ErrNotFound
		}
		return err
	}

	inp := input.(*UpdateInput)

	var selects []interface{}
	if inp.Age != nil {
		prof.Age = inp.Age
		selects = append(selects, "age")
	}
	if inp.Height != nil {
		prof.Height = inp.Height
		selects = append(selects, "height")
	}
	if inp.Address != nil && *inp.Address != "" {
		prof.Address = *inp.Address
		selects = append(selects, "address")
	}

	if len(selects) > 0 {
		if err = s.profileRepository.Update(ctx, prof, selects, `user_id = ?`, userID); err != nil {
			if apperr.Is(err, apperr.ErrDBDuplicated) {
				return profile.ErrExists
			}
			return err
		}
	}

	return nil
}

type UpdateProfileInput struct {
	Age     *int
	Height  *int
	Address *string
}

func (s Service) UpdateProfile(ctx context.Context, userID int64, input any) error {
	prof, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return profile.ErrNotFound
		}
		return err
	}

	inp := input.(*UpdateProfileInput)

	var selects []interface{}
	if inp.Age != nil {
		if *inp.Age == 0 {
			prof.Age = nil
		} else {
			prof.Age = inp.Age
		}

		selects = append(selects, "age")
	}
	if inp.Height != nil {
		if *inp.Height == 0 {
			prof.Height = nil
		} else {
			prof.Height = inp.Height
		}
		selects = append(selects, "height")
	}
	if inp.Address != nil && *inp.Address != "" {
		prof.Address = *inp.Address
		selects = append(selects, "address")
	}

	if len(selects) > 0 {
		if err = s.profileRepository.Update(ctx, prof, selects, `user_id = ?`, userID); err != nil {
			if apperr.Is(err, apperr.ErrDBDuplicated) {
				return profile.ErrExists
			}
			return err
		}
	}

	return nil
}

func (s Service) Delete(ctx context.Context, userID int64) error {
	_, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return profile.ErrNotFound
		}
		return err
	}

	if err := s.profileRepository.Delete(ctx, `user_id = ?`, userID); err != nil {
		return err
	}

	return nil
}
