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
	getByIDFunc    func(ctx context.Context, id int64) (*domain.User, error)
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

func (m *mockUserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.User{
		ID:           id,
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func TestNewUserUsecase(t *testing.T) {
	mockRepo := &mockUserRepository{}
	mockProfileRepo := &mockUserProfileRepository{}
	usecase := NewUserUsecase(mockRepo, mockProfileRepo)

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
			mockProfileRepo := &mockUserProfileRepository{}
			usecase := NewUserUsecase(mockRepo, mockProfileRepo)

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
			mockProfileRepo := &mockUserProfileRepository{}
			usecase := NewUserUsecase(mockRepo, mockProfileRepo)

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

func TestGetUserByID(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name        string
		id          int64
		mockGetByID func(ctx context.Context, id int64) (*domain.User, error)
		wantID      int64
		wantUser    bool
		wantErr     bool
	}{
		{
			name: "user found",
			id:   123,
			mockGetByID: func(ctx context.Context, id int64) (*domain.User, error) {
				return &domain.User{
					ID:           id,
					Email:        "test@example.com",
					PasswordHash: "hashed_password",
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantID:   123,
			wantUser: true,
			wantErr:  false,
		},
		{
			name: "user not found",
			id:   999,
			mockGetByID: func(ctx context.Context, id int64) (*domain.User, error) {
				return nil, nil
			},
			wantID:   0,
			wantUser: false,
			wantErr:  false,
		},
		{
			name: "repository error",
			id:   456,
			mockGetByID: func(ctx context.Context, id int64) (*domain.User, error) {
				return nil, errors.New("database error")
			},
			wantID:   0,
			wantUser: false,
			wantErr:  true,
		},
		{
			name: "user with ID 1",
			id:   1,
			mockGetByID: func(ctx context.Context, id int64) (*domain.User, error) {
				return &domain.User{
					ID:           1,
					Email:        "admin@example.com",
					PasswordHash: "hashed_password",
					CreatedAt:    now,
					UpdatedAt:    now,
				}, nil
			},
			wantID:   1,
			wantUser: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{
				getByIDFunc: tt.mockGetByID,
			}
			mockProfileRepo := &mockUserProfileRepository{}
			usecase := NewUserUsecase(mockRepo, mockProfileRepo)

			user, err := usecase.GetUserByID(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantUser {
				if user == nil {
					t.Error("GetUserByID() returned nil user")
					return
				}

				if user.ID != tt.wantID {
					t.Errorf("GetUserByID() id = %v, want %v", user.ID, tt.wantID)
				}
			} else if !tt.wantErr && user != nil {
				t.Error("GetUserByID() expected nil user")
			}
		})
	}
}

func TestCheckUserProfileExists(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		userID             int64
		mockExistsByUserID func(ctx context.Context, userID int64) (bool, error)
		wantExists         bool
		wantErr            bool
	}{
		{
			name:   "profile exists",
			userID: 123,
			mockExistsByUserID: func(ctx context.Context, userID int64) (bool, error) {
				return true, nil
			},
			wantExists: true,
			wantErr:    false,
		},
		{
			name:   "profile does not exist",
			userID: 456,
			mockExistsByUserID: func(ctx context.Context, userID int64) (bool, error) {
				return false, nil
			},
			wantExists: false,
			wantErr:    false,
		},
		{
			name:   "repository error",
			userID: 789,
			mockExistsByUserID: func(ctx context.Context, userID int64) (bool, error) {
				return false, errors.New("database error")
			},
			wantExists: false,
			wantErr:    true,
		},
		{
			name:   "user ID is 0",
			userID: 0,
			mockExistsByUserID: func(ctx context.Context, userID int64) (bool, error) {
				return false, nil
			},
			wantExists: false,
			wantErr:    false,
		},
		{
			name:   "negative user ID",
			userID: -1,
			mockExistsByUserID: func(ctx context.Context, userID int64) (bool, error) {
				return false, nil
			},
			wantExists: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			mockProfileRepo := &mockUserProfileRepository{
				existsByUserIDFunc: tt.mockExistsByUserID,
			}
			usecase := NewUserUsecase(mockRepo, mockProfileRepo)

			exists, err := usecase.CheckUserProfileExists(ctx, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckUserProfileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if exists != tt.wantExists {
				t.Errorf("CheckUserProfileExists() exists = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}
