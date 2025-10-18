package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/keu-5/muzee/backend/internal/domain"
)

// Mock UserRepository
type mockUserRepository struct {
	createFunc     func(ctx context.Context, email, passwordHash string) (*domain.User, error)
	getByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
}

func (m *mockUserRepository) Create(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, email, passwordHash)
	}
	return &domain.User{
		ID:           1,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return nil, nil
}

func TestNewUserUsecase(t *testing.T) {
	mockRepo := &mockUserRepository{}
	usecase := NewUserUsecase(mockRepo)

	if usecase == nil {
		t.Fatal("Expected usecase to be non-nil")
	}
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name         string
		email        string
		passwordHash string
		mockCreate   func(ctx context.Context, email, passwordHash string) (*domain.User, error)
		wantEmail    string
		wantErr      bool
	}{
		{
			name:         "successful user creation",
			email:        "test@example.com",
			passwordHash: "hashed_password",
			mockCreate: func(ctx context.Context, email, passwordHash string) (*domain.User, error) {
				return &domain.User{
					ID:           1,
					Email:        email,
					PasswordHash: passwordHash,
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name:         "email with uppercase",
			email:        "TEST@EXAMPLE.COM",
			passwordHash: "hashed_password",
			mockCreate: func(ctx context.Context, email, passwordHash string) (*domain.User, error) {
				return &domain.User{
					ID:           2,
					Email:        email,
					PasswordHash: passwordHash,
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name:         "email with spaces",
			email:        "  test@example.com  ",
			passwordHash: "hashed_password",
			mockCreate: func(ctx context.Context, email, passwordHash string) (*domain.User, error) {
				return &domain.User{
					ID:           3,
					Email:        email,
					PasswordHash: passwordHash,
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name:         "email with uppercase and spaces",
			email:        "  TEST@EXAMPLE.COM  ",
			passwordHash: "hashed_password",
			mockCreate: func(ctx context.Context, email, passwordHash string) (*domain.User, error) {
				return &domain.User{
					ID:           4,
					Email:        email,
					PasswordHash: passwordHash,
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantEmail: "test@example.com",
			wantErr:   false,
		},
		{
			name:         "repository error",
			email:        "error@example.com",
			passwordHash: "hashed_password",
			mockCreate: func(ctx context.Context, email, passwordHash string) (*domain.User, error) {
				return nil, errors.New("database error")
			},
			wantEmail: "",
			wantErr:   true,
		},
		{
			name:         "duplicate email error",
			email:        "duplicate@example.com",
			passwordHash: "hashed_password",
			mockCreate: func(ctx context.Context, email, passwordHash string) (*domain.User, error) {
				return nil, errors.New("unique constraint violation")
			},
			wantEmail: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{
				createFunc: tt.mockCreate,
			}
			usecase := NewUserUsecase(mockRepo)

			user, err := usecase.CreateUser(ctx, tt.email, tt.passwordHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if user == nil {
					t.Error("CreateUser() returned nil user")
					return
				}

				if user.Email != tt.wantEmail {
					t.Errorf("CreateUser() email = %v, want %v", user.Email, tt.wantEmail)
				}

				if user.PasswordHash != tt.passwordHash {
					t.Errorf("CreateUser() passwordHash = %v, want %v", user.PasswordHash, tt.passwordHash)
				}
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name          string
		email         string
		mockGetByEmail func(ctx context.Context, email string) (*domain.User, error)
		wantEmail     string
		wantUser      bool
		wantErr       bool
	}{
		{
			name:  "user found",
			email: "test@example.com",
			mockGetByEmail: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{
					ID:           1,
					Email:        email,
					PasswordHash: "hashed_password",
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantEmail: "test@example.com",
			wantUser:  true,
			wantErr:   false,
		},
		{
			name:  "user not found",
			email: "notfound@example.com",
			mockGetByEmail: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, nil
			},
			wantEmail: "",
			wantUser:  false,
			wantErr:   false,
		},
		{
			name:  "email with uppercase",
			email: "TEST@EXAMPLE.COM",
			mockGetByEmail: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{
					ID:           2,
					Email:        email,
					PasswordHash: "hashed_password",
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantEmail: "test@example.com",
			wantUser:  true,
			wantErr:   false,
		},
		{
			name:  "email with spaces",
			email: "  test@example.com  ",
			mockGetByEmail: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{
					ID:           3,
					Email:        email,
					PasswordHash: "hashed_password",
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantEmail: "test@example.com",
			wantUser:  true,
			wantErr:   false,
		},
		{
			name:  "email with uppercase and spaces",
			email: "  TEST@EXAMPLE.COM  ",
			mockGetByEmail: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{
					ID:           4,
					Email:        email,
					PasswordHash: "hashed_password",
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantEmail: "test@example.com",
			wantUser:  true,
			wantErr:   false,
		},
		{
			name:  "repository error",
			email: "error@example.com",
			mockGetByEmail: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, errors.New("database error")
			},
			wantEmail: "",
			wantUser:  false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{
				getByEmailFunc: tt.mockGetByEmail,
			}
			usecase := NewUserUsecase(mockRepo)

			user, err := usecase.GetUserByEmail(ctx, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantUser {
				if user == nil {
					t.Error("GetUserByEmail() returned nil user")
					return
				}

				if user.Email != tt.wantEmail {
					t.Errorf("GetUserByEmail() email = %v, want %v", user.Email, tt.wantEmail)
				}
			} else if !tt.wantErr && user != nil {
				t.Error("GetUserByEmail() expected nil user")
			}
		})
	}
}
