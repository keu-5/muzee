package usecase

import (
	"context"
	"strings"

	"github.com/keu-5/muzee/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase interface {
	HashPassword(password string) (string, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
}

type authUsecase struct {
	userRepo repository.UserRepository
}

func NewAuthUsecase(userRepo repository.UserRepository) AuthUsecase {
	return &authUsecase{userRepo: userRepo}
}

func (a *authUsecase) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (a *authUsecase) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	return user != nil, nil
}
