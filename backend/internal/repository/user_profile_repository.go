package repository

import (
	"context"

	"github.com/keu-5/muzee/backend/ent"
	"github.com/keu-5/muzee/backend/internal/domain"
)

type UserProfileRepository interface {
	Create(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error)
}

type userProfileRepository struct {
	client *ent.Client
}

func NewUserProfileRepository(client *ent.Client) UserProfileRepository {
	return &userProfileRepository{client: client}
}

func (r *userProfileRepository) Create(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
	profile, err := r.client.UserProfile.Create().
		SetUserID(userID).
		SetName(name).
		SetUsername(username).
		SetNillableIconPath(iconPath).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.UserProfile{
		ID:        profile.ID,
		Name:      profile.Name,
		Username:  profile.Username,
		IconPath:  profile.IconPath,
		CreatedAt: profile.CreatedAt,
		UpdatedAt: profile.UpdatedAt,
	}, nil
}
