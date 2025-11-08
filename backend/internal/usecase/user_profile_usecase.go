package usecase

import (
	"context"
	"mime/multipart"

	"github.com/keu-5/muzee/backend/internal/domain"
	"github.com/keu-5/muzee/backend/internal/helper"
	"github.com/keu-5/muzee/backend/internal/repository"
)

type UserProfileUsecase interface {
	CreateUserProfile(ctx context.Context, userID int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error)
	IsUsernameAvailable(ctx context.Context, username string) (bool, error)
}

type userProfileUsecase struct {
	userProfileRepo repository.UserProfileRepository
	fileHelper      *helper.FileHelper
}

func NewUserProfileUsecase(userProfileRepo repository.UserProfileRepository, fileHelper *helper.FileHelper) UserProfileUsecase {
	return &userProfileUsecase{
		userProfileRepo: userProfileRepo,
		fileHelper:      fileHelper,
	}
}

func (u *userProfileUsecase) CreateUserProfile(ctx context.Context, userID int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error) {
	var iconPath *string

	// 画像ファイルが提供されている場合、保存処理
	if iconFile != nil {
		savedPath, err := u.fileHelper.SaveImageFile(iconFile, "profiles")
		if err != nil {
			return nil, err
		}
		iconPath = &savedPath
	}

	// ユーザープロフィールをDBに保存
	userProfile, err := u.userProfileRepo.Create(ctx, userID, name, username, iconPath)
	if err != nil {
		// 画像保存に成功していた場合、ロールバックとして画像を削除
		if iconPath != nil {
			_ = u.fileHelper.DeleteImageFile(*iconPath)
		}
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
