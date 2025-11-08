package usecase

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/domain"
	"github.com/keu-5/muzee/backend/internal/infrastructure"
	"github.com/keu-5/muzee/backend/internal/repository"
)

const userIconsFolder = "user-icons"

type UserProfileUsecase interface {
	CreateUserProfile(ctx context.Context, userID int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error)
	IsUsernameAvailable(ctx context.Context, username string) (bool, error)
}

type userProfileUsecase struct {
	userProfileRepo repository.UserProfileRepository
	storageService  *infrastructure.StorageService
	cfg             *config.Config
}

func NewUserProfileUsecase(userProfileRepo repository.UserProfileRepository, storageService *infrastructure.StorageService, cfg *config.Config) UserProfileUsecase {
	return &userProfileUsecase{
		userProfileRepo: userProfileRepo,
		storageService:  storageService,
		cfg:             cfg,
	}
}

func (u *userProfileUsecase) CreateUserProfile(ctx context.Context, userID int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error) {
	var iconPath *string

	if iconFile != nil {
		// Generate unique object name: user-icons/user_{userID}/{randomhex}.{ext}
		prefix := fmt.Sprintf("%s/user_%d", userIconsFolder, userID)
		objectName := u.storageService.GenerateUniqueObjectName(prefix, iconFile.Filename)

		// Upload user icons to public bucket
		if err := u.storageService.UploadFile(ctx, u.cfg.S3PublicBucket, objectName, iconFile); err != nil {
			return nil, err
		}

		iconPath = &objectName
	}

	userProfile, err := u.userProfileRepo.Create(ctx, userID, name, username, iconPath)
	if err != nil {
		return nil, err
	}
	return userProfile, nil
}

func (u *userProfileUsecase) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	exists, err := u.userProfileRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return false, err
	}
	return !exists, nil
}
