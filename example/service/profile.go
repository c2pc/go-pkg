package service

import (
	"context"
	"github.com/c2pc/go-pkg/v2/auth/profile"
	"github.com/c2pc/go-pkg/v2/example/i18n"
	"github.com/c2pc/go-pkg/v2/example/model"
	"github.com/c2pc/go-pkg/v2/example/repository"
	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	model2 "github.com/c2pc/go-pkg/v2/utils/model"
	"gorm.io/gorm"
)

var (
	ErrProfileNotFound = apperr.New("profile_not_found", apperr.WithTextTranslate(i18n.ErrProfileNotFound), apperr.WithCode(code.NotFound))
	ErrProfileExists   = apperr.New("profile_exists_error", apperr.WithTextTranslate(i18n.ErrProfileExists), apperr.WithCode(code.InvalidArgument))
)

type ProfileService[Model model.Profile, CreateInput ProfileCreateInput, UpdateInput ProfileUpdateInput, UpdateProfileInput ProfileUpdateProfileInput] struct {
	profileRepository repository.IProfileRepository
}

func NewProfileService[Model model.Profile, CreateInput ProfileCreateInput, UpdateInput ProfileUpdateInput, UpdateProfileInput ProfileUpdateProfileInput](
	profileRepository repository.IProfileRepository,
) ProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	return ProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]{
		profileRepository: profileRepository,
	}
}

func (s ProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Trx(db *gorm.DB) profile.IProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput] {
	s.profileRepository = s.profileRepository.Trx(db)

	return s
}

func (s ProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]) GetById(ctx context.Context, userID int) (*Model, error) {
	prof, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}

	m := Model(*prof)

	return &m, nil
}

func (s ProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]) GetByIds(ctx context.Context, userID ...int) ([]Model, error) {
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
	Login string
	Name  string
}

func (s ProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Create(ctx context.Context, userID int, input CreateInput) (*Model, error) {
	inp := ProfileCreateInput(input)

	prof, err := s.profileRepository.Create(ctx, &model.Profile{
		Login:  inp.Login,
		Name:   inp.Name,
		UserID: userID,
	}, "id")
	if err != nil {
		if apperr.Is(err, apperr.ErrDBDuplicated) {
			return nil, ErrProfileExists
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
	Login *string
	Name  *string
}

func (s ProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Update(ctx context.Context, userID int, input UpdateInput) error {
	prof, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrProfileNotFound
		}
		return err
	}

	inp := ProfileUpdateInput(input)

	var selects []interface{}
	if inp.Login != nil && *inp.Login != "" {
		prof.Login = *inp.Login
		selects = append(selects, "login")
	}
	if inp.Name != nil && *inp.Name != "" {
		prof.Name = *inp.Name
		selects = append(selects, "name")
	}

	if len(selects) > 0 {
		if err = s.profileRepository.Update(ctx, prof, selects, `user_id = ?`, userID); err != nil {
			if apperr.Is(err, apperr.ErrDBDuplicated) {
				return ErrProfileExists
			}
			return err
		}
	}

	return nil
}

type ProfileUpdateProfileInput struct {
	Login *string
	Name  *string
}

func (s ProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]) UpdateProfile(ctx context.Context, userID int, input UpdateProfileInput) error {
	prof, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrProfileNotFound
		}
		return err
	}

	inp := ProfileUpdateProfileInput(input)

	var selects []interface{}
	if inp.Login != nil && *inp.Login != "" {
		prof.Login = *inp.Login
		selects = append(selects, "login")
	}
	if inp.Name != nil && *inp.Name != "" {
		prof.Name = *inp.Name
		selects = append(selects, "name")
	}

	if len(selects) > 0 {
		if err = s.profileRepository.Update(ctx, prof, selects, `user_id = ?`, userID); err != nil {
			if apperr.Is(err, apperr.ErrDBDuplicated) {
				return ErrProfileExists
			}
			return err
		}
	}

	return nil
}

func (s ProfileService[Model, CreateInput, UpdateInput, UpdateProfileInput]) Delete(ctx context.Context, userID int) error {
	_, err := s.profileRepository.Find(ctx, `user_id = ?`, userID)
	if err != nil {
		if apperr.Is(err, apperr.ErrDBRecordNotFound) {
			return ErrProfileNotFound
		}
		return err
	}

	if err := s.profileRepository.Delete(ctx, `user_id = ?`, userID); err != nil {
		return err
	}

	return nil
}
