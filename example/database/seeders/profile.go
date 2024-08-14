package seeders

import (
	"context"
	"github.com/c2pc/go-pkg/v2/example/model"
	"github.com/c2pc/go-pkg/v2/example/repository"
)

func ProfileSeeder(ctx context.Context, profileRepository repository.IProfileRepository, adminID int) (*model.Profile, error) {
	login := "admin"
	name := "Admin"

	admin, err := profileRepository.FirstOrCreate(ctx, &model.Profile{
		UserID: adminID,
		Login:  login,
		Name:   name,
	}, "", `user_id = ?`, adminID)
	if err != nil {
		return nil, err
	}

	return admin, nil
}
