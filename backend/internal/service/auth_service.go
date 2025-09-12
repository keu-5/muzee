package service

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/db"
	"github.com/keu-5/muzee/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
	config   *config.Config
}

func NewAuthService(config *config.Config) *AuthService {
	return &AuthService{
		userRepo: repository.NewUserRepository(),
		config:   config,
	}
}

func (s *AuthService) AuthenticateUser(username, password string) (db.User, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return db.User{}, err
	}
	if user.ID == 0 {
		return db.User{}, nil
	}
	
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return db.User{}, nil
	}
	
	return user, nil
}

func (s *AuthService) CreateUser(username, email, password string) (db.User, error) {
	// ユーザー名の重複チェック
	existingUser, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return db.User{}, err
	}
	if existingUser.ID != 0 {
		return db.User{}, errors.New("username already exists")
	}

	// メールアドレスの重複チェック
	existingUser, err = s.userRepo.GetUserByEmail(email)
	if err != nil {
		return db.User{}, err
	}
	if existingUser.ID != 0 {
		return db.User{}, errors.New("email already exists")
	}

	// パスワードハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, err
	}

	params := db.CreateUserParams{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	user, err := s.userRepo.CreateUser(params)
	if err != nil {
		return db.User{}, err
	}

	return user, nil
}

func (s *AuthService) GenerateToken(userID int32) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.Itoa(int(userID)),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}
