package usecase

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase interface {
	HashPassword(password string) (string, error)
	CheckEmailExists(ctx context.Context, email string) (bool, error)
}

type authUsecase struct {
	userUC UserUsecase
}

func NewAuthUsecase(userUC UserUsecase) AuthUsecase {
	return &authUsecase{userUC: userUC}
}

func (a *authUsecase) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (a *authUsecase) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	user, err := a.userUC.GetUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	return user != nil, nil
}
