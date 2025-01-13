package profile

import (
	"context"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"gorm.io/gorm"
)

var (
	ErrNotFound = apperr.New("profile_not_found", apperr.WithTextTranslate(ErrNotFoundTranslate), apperr.WithCode(code.NotFound))
	ErrExists   = apperr.New("profile_exists_error", apperr.WithTextTranslate(ErrExistsTranslate), apperr.WithCode(code.InvalidArgument))
)

type Service[Model Profile, CreateInput ProfileCreateInput, UpdateInput ProfileUpdateInput, UpdateProfileInput ProfileUpdateProfileInput] struct {
	profileRepository IRepository
}

func NewService[Model Profile, CreateInput ProfileCreateInput, UpdateInput ProfileUpdateInput, UpdateProfileInput ProfileUpdateProfileInput](
	profileRepository IRepository,
) Service[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	return Service[Model, CreateInput, UpdateInput, UpdateProfileInput]{
		profileRepository: profileRepository,
	}
}

func (s Service[Model, CreateInput, UpdateInput, UpdateProfileInput]) Trx(db *gorm.DB) profile.IProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	s.profileRepository = s.profileRepository.Trx(db)

	return s
}

func (s Service[Model, CreateInput, UpdateInput, UpdateProfileInput]) GetById(ctx context.Context, userID int) (*Model, error) {
	prof, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	m := Model(*prof)

	return &m, nil
}

func (s Service[Model, CreateInput, UpdateInput, UpdateProfileInput]) GetByIds(ctx context.Context, userID ...int) ([]Model, error) {
	profs, err := s.profileRepository.List(ctx, &model2.Filter{}, `user_id IN (?)`, userID)
	if err != nil {
		return nil, err
	}

	m := make([]Model, len(profs))
	for i, prof := range profs {
		m[i] = Model(prof)
	}

	return m, nil
}

type ProfileCreateInput struct {
	Age     *int
	Height  *int
	Address string
}

func (s Service[Model, CreateInput, UpdateInput, UpdateProfileInput]) Create(ctx context.Context, userID int, input CreateInput) (*Model, error) {
	inp := ProfileCreateInput(input)

	prof, err := s.profileRepository.Create(ctx, &Profile{
		Age:     inp.Age,
		Height:  inp.Height,
		Address: inp.Address,
		UserID:  userID,
	}, "id")
	if err != nil {
		if apperr.Is(err, apperr.ErrDBDuplicated) {
			return nil, ErrExists
		}
		return nil, err
	}

	prof, err = s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		return nil, err
	}

	m := Model(*prof)

	return &m, nil
}

type ProfileUpdateInput struct {
	Age     *int
	Height  *int
	Address *string
}

func (s Service[Model, CreateInput, UpdateInput, UpdateProfileInput]) Update(ctx context.Context, userID int, input UpdateInput) error {
	prof, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	inp := ProfileUpdateInput(input)

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
				return ErrExists
			}
			return err
		}
	}

	return nil
}

type ProfileUpdateProfileInput struct {
	Age     *int
	Height  *int
	Address *string
}

func (s Service[Model, CreateInput, UpdateInput, UpdateProfileInput]) UpdateProfile(ctx context.Context, userID int, input UpdateProfileInput) error {
	prof, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	inp := ProfileUpdateProfileInput(input)

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
				return ErrExists
			}
			return err
		}
	}

	return nil
}

func (s Service[Model, CreateInput, UpdateInput, UpdateProfileInput]) Delete(ctx context.Context, userID int) error {
	_, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	if err := s.profileRepository.Delete(ctx, `user_id = ?`, userID); err != nil {
		return err
	}

	return nil
}
