package usecase

import (
	"context"

	"github.com/keu-5/muzee/backend/internal/domain"
	"github.com/keu-5/muzee/backend/internal/repository"
)

type UserProfileUsecase interface {
	CreateUserProfile(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error)
}

type userProfileUsecase struct {
	userProfileRepo repository.UserProfileRepository
}

func NewUserProfileUsecase(userProfileRepo repository.UserProfileRepository) UserProfileUsecase {
	return &userProfileUsecase{
		userProfileRepo: userProfileRepo,
	}
}

func (u *userProfileUsecase) CreateUserProfile(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
	userProfile, err := u.userProfileRepo.Create(ctx, userID, name, username, iconPath)
	if err != nil {
		return nil, err
	}
	return userProfile, nil
}
