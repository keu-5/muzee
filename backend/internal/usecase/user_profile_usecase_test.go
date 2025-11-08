package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/keu-5/muzee/backend/internal/domain"
	"github.com/keu-5/muzee/backend/internal/helper"
)

// Mock UserProfileRepository
type mockUserProfileRepository struct {
	createFunc             func(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error)
	existsByUserIDFunc     func(ctx context.Context, userID int64) (bool, error)
	existsByUsernameFunc   func(ctx context.Context, username string) (bool, error)
}

func (m *mockUserProfileRepository) Create(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, userID, name, username, iconPath)
	}
	return &domain.UserProfile{
		ID:        1,
		UserID:    userID,
		Name:      name,
		Username:  username,
		IconPath:  iconPath,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockUserProfileRepository) ExistsByUserID(ctx context.Context, userID int64) (bool, error) {
	if m.existsByUserIDFunc != nil {
		return m.existsByUserIDFunc(ctx, userID)
	}
	return false, nil
}

func (m *mockUserProfileRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	if m.existsByUsernameFunc != nil {
		return m.existsByUsernameFunc(ctx, username)
	}
	return false, nil
}

func TestNewUserProfileUsecase(t *testing.T) {
	mockRepo := &mockUserProfileRepository{}
	mockFileHelper := &helper.FileHelper{}
	usecase := NewUserProfileUsecase(mockRepo, mockFileHelper)

	if usecase == nil {
		t.Fatal("Expected usecase to be non-nil")
	}
}

func TestCreateUserProfile(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	tests := []struct {
		name         string
		userID       int64
		profileName  string
		username     string
		iconPath     *string
		mockCreate   func(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error)
		wantName     string
		wantUsername string
		wantErr      bool
	}{
		{
			name:        "successful user profile creation",
			userID:      123,
			profileName: "Test User",
			username:    "testuser",
			iconPath:    nil,
			mockCreate: func(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
				return &domain.UserProfile{
					ID:        1,
					UserID:    userID,
					Name:      name,
					Username:  username,
					IconPath:  iconPath,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			},
			wantName:     "Test User",
			wantUsername: "testuser",
			wantErr:      false,
		},
		{
			name:        "user profile creation with icon path",
			userID:      456,
			profileName: "Another User",
			username:    "anotheruser",
			iconPath:    stringPtr("https://example.com/icon.png"),
			mockCreate: func(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
				return &domain.UserProfile{
					ID:        2,
					UserID:    userID,
					Name:      name,
					Username:  username,
					IconPath:  iconPath,
					CreatedAt: now,
					UpdatedAt: now,
				}, nil
			},
			wantName:     "Another User",
			wantUsername: "anotheruser",
			wantErr:      false,
		},
		{
			name:        "repository error",
			userID:      789,
			profileName: "Error User",
			username:    "erroruser",
			iconPath:    nil,
			mockCreate: func(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
				return nil, errors.New("database error")
			},
			wantName:     "",
			wantUsername: "",
			wantErr:      true,
		},
		{
			name:        "duplicate username error",
			userID:      999,
			profileName: "Duplicate User",
			username:    "duplicate",
			iconPath:    nil,
			mockCreate: func(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
				return nil, errors.New("unique constraint violation")
			},
			wantName:     "",
			wantUsername: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserProfileRepository{
				createFunc: tt.mockCreate,
			}
			mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
			usecase := NewUserProfileUsecase(mockRepo, mockFileHelper)

			// iconFileはnilとして渡す（ファイルアップロードテストは別途実施）
			profile, err := usecase.CreateUserProfile(ctx, tt.userID, tt.profileName, tt.username, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUserProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if profile == nil {
					t.Error("CreateUserProfile() returned nil profile")
					return
				}

				if profile.Name != tt.wantName {
					t.Errorf("CreateUserProfile() name = %v, want %v", profile.Name, tt.wantName)
				}

				if profile.Username != tt.wantUsername {
					t.Errorf("CreateUserProfile() username = %v, want %v", profile.Username, tt.wantUsername)
				}

				if profile.UserID != tt.userID {
					t.Errorf("CreateUserProfile() userID = %v, want %v", profile.UserID, tt.userID)
				}

				// iconFileはnilで渡しているため、iconPathは常にnilになる
				if profile.IconPath != nil {
					t.Errorf("CreateUserProfile() iconPath = %v, want nil (iconFile was nil)", profile.IconPath)
				}
			}
		})
	}
}

func TestIsUsernameAvailable(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name               string
		username           string
		mockExistsByUsername func(ctx context.Context, username string) (bool, error)
		wantAvailable      bool
		wantErr            bool
	}{
		{
			name:     "username is available",
			username: "newuser",
			mockExistsByUsername: func(ctx context.Context, username string) (bool, error) {
				return false, nil
			},
			wantAvailable: true,
			wantErr:       false,
		},
		{
			name:     "username is not available",
			username: "existinguser",
			mockExistsByUsername: func(ctx context.Context, username string) (bool, error) {
				return true, nil
			},
			wantAvailable: false,
			wantErr:       false,
		},
		{
			name:     "repository error",
			username: "testuser",
			mockExistsByUsername: func(ctx context.Context, username string) (bool, error) {
				return false, errors.New("database error")
			},
			wantAvailable: false,
			wantErr:       true,
		},
		{
			name:     "empty username check",
			username: "",
			mockExistsByUsername: func(ctx context.Context, username string) (bool, error) {
				return false, nil
			},
			wantAvailable: true,
			wantErr:       false,
		},
		{
			name:     "username with special characters",
			username: "user_name123",
			mockExistsByUsername: func(ctx context.Context, username string) (bool, error) {
				return false, nil
			},
			wantAvailable: true,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserProfileRepository{
				existsByUsernameFunc: tt.mockExistsByUsername,
			}
			mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
			usecase := NewUserProfileUsecase(mockRepo, mockFileHelper)

			available, err := usecase.IsUsernameAvailable(ctx, tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsUsernameAvailable() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if available != tt.wantAvailable {
				t.Errorf("IsUsernameAvailable() available = %v, want %v", available, tt.wantAvailable)
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
