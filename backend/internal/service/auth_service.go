package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/ent"
	"github.com/keu-5/muzee/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
	config   *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, config *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		config:   config,
	}
}

func (s *AuthService) AuthenticateUser(ctx context.Context, username, password string) (*ent.User, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, nil
	}
	
	return user, nil
}

func (s *AuthService) CreateUser(ctx context.Context, username, email, password string) (*ent.User, error) {
	// ユーザー名の重複チェック
	existingUser, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// メールアドレスの重複チェック
	existingUser, err = s.userRepo.GetUserByEmail(ctx, email)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// パスワードハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.CreateUser(ctx, username, email, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.Itoa(userID),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}
