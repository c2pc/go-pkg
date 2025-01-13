package seeders

import (
	"context"
	"github.com/c2pc/go-pkg/v2/example/profile"
)

func ProfileSeeder(ctx context.Context, profileRepository profile.IRepository, adminID int) (*profile.Profile, error) {
	admin, err := profileRepository.FirstOrCreate(ctx, &profile.Profile{
		UserID:  adminID,
		Age:     nil,
		Height:  nil,
		Address: "Moscow",
	}, "", `user_id = ?`, adminID)
	if err != nil {
		return nil, err
	}

	return admin, nil
}
