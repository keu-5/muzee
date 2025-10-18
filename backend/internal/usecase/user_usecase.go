package usecase

import (
	"context"
	"strings"

	"github.com/keu-5/muzee/backend/internal/domain"
	"github.com/keu-5/muzee/backend/internal/repository"
)

type UserUsecase interface {
	CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type userUsecase struct {
	userRepo repository.UserRepository
}

func NewUserUsecase(userRepo repository.UserRepository) UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := u.userRepo.Create(ctx, email, passwordHash)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *userUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	return u.userRepo.GetByEmail(ctx, email)
}
