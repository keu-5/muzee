package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/keu-5/muzee/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// Mock UserRepository
type mockUserRepository struct {
	getByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
	createFunc     func(ctx context.Context, email, passwordHash string) (*domain.User, error)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserRepository) Create(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, email, passwordHash)
	}
	return nil, nil
}

func TestNewAuthUsecase(t *testing.T) {
	mockRepo := &mockUserRepository{}
	usecase := NewAuthUsecase(mockRepo)

	if usecase == nil {
		t.Fatal("Expected usecase to be non-nil")
	}
}

func TestHashPassword(t *testing.T) {
	mockRepo := &mockUserRepository{}
	usecase := NewAuthUsecase(mockRepo)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "this_is_a_very_long_password_with_many_characters_1234567890",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := usecase.HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify that the hash is not empty
				if hash == "" {
					t.Error("HashPassword() returned empty hash")
				}

				// Verify that the hash is different from the original password
				if hash == tt.password {
					t.Error("HashPassword() returned the original password")
				}

				// Verify that the hash can be used to compare with the original password
				err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(tt.password))
				if err != nil {
					t.Errorf("Generated hash cannot be verified with original password: %v", err)
				}
			}
		})
	}
}

func TestCheckEmailExists(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name          string
		email         string
		mockGetByEmail func(ctx context.Context, email string) (*domain.User, error)
		wantExists    bool
		wantErr       bool
	}{
		{
			name:  "email exists",
			email: "test@example.com",
			mockGetByEmail: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{
					ID:           1,
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:  "email does not exist",
			email: "nonexistent@example.com",
			mockGetByEmail: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, nil
			},
			wantExists: false,
			wantErr:    false,
		},
		{
			name:  "email with uppercase and spaces",
			email: "  TEST@EXAMPLE.COM  ",
			mockGetByEmail: func(ctx context.Context, email string) (*domain.User, error) {
				// Should receive normalized email (lowercase, trimmed)
				if email != "test@example.com" {
					t.Errorf("Expected normalized email 'test@example.com', got '%s'", email)
				}
				return &domain.User{
					ID:           1,
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:  "repository error",
			email: "error@example.com",
			mockGetByEmail: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, errors.New("database error")
			},
			wantExists: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{
				getByEmailFunc: tt.mockGetByEmail,
			}
			usecase := NewAuthUsecase(mockRepo)

			exists, err := usecase.CheckEmailExists(ctx, tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckEmailExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if exists != tt.wantExists {
				t.Errorf("CheckEmailExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}
