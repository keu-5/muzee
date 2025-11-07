package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/domain"
	"github.com/keu-5/muzee/backend/internal/helper"
	"github.com/keu-5/muzee/backend/internal/interface/middleware"
	"github.com/keu-5/muzee/backend/internal/util"
	"github.com/stretchr/testify/assert"
)

// Mock UserProfileUsecase
type mockUserProfileUsecase struct {
	createUserProfileFunc func(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error)
}

func (m *mockUserProfileUsecase) CreateUserProfile(ctx context.Context, userID int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
	if m.createUserProfileFunc != nil {
		return m.createUserProfileFunc(ctx, userID, name, username, iconPath)
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

func setupTestUserProfileApp(handler *UserProfileHandler, jwtSecret string) *fiber.App {
	app := fiber.New()
	app.Post("/api/v1/users/me/profile", middleware.AuthMiddleware(jwtSecret), handler.CreateMyProfile)
	return app
}

func TestNewUserProfileHandler(t *testing.T) {
	mockUserProfile := &mockUserProfileUsecase{}
	handler := NewUserProfileHandler(mockUserProfile)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.userProfileUC)
}

func TestCreateMyProfile_Success(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{
		createUserProfileFunc: func(ctx context.Context, uid int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
			return &domain.UserProfile{
				ID:        1,
				UserID:    uid,
				Name:      name,
				Username:  username,
				IconPath:  iconPath,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	handler := NewUserProfileHandler(mockUserProfile)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Create request
	reqBody := CreateMyProfileRequest{
		Name:     "Test User",
		Username: "testuser",
		IconPath: "https://example.com/icon.png",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 201, resp.StatusCode)

	var response CreateMyProfileResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "プロフィール")
	assert.Equal(t, int64(1), response.UserProfile.ID)
	assert.Equal(t, "Test User", response.UserProfile.Name)
	assert.Equal(t, "testuser", response.UserProfile.Username)
	assert.Equal(t, "https://example.com/icon.png", response.UserProfile.IconPath)
}

func TestCreateMyProfile_Success_WithoutIconPath(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{
		createUserProfileFunc: func(ctx context.Context, uid int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
			return &domain.UserProfile{
				ID:        1,
				UserID:    uid,
				Name:      name,
				Username:  username,
				IconPath:  iconPath,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	handler := NewUserProfileHandler(mockUserProfile)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Create request without icon_path
	reqBody := CreateMyProfileRequest{
		Name:     "Test User",
		Username: "testuser",
		IconPath: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 201, resp.StatusCode)

	var response CreateMyProfileResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test User", response.UserProfile.Name)
	assert.Equal(t, "testuser", response.UserProfile.Username)
	assert.Equal(t, "", response.UserProfile.IconPath)
}

func TestCreateMyProfile_InvalidJSON(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{}
	handler := NewUserProfileHandler(mockUserProfile)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp map[string]string
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "bad_request", errResp["error"])
}

func TestCreateMyProfile_ValidationError_MissingName(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{}
	handler := NewUserProfileHandler(mockUserProfile)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Missing name
	reqBody := CreateMyProfileRequest{
		Name:     "",
		Username: "testuser",
		IconPath: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "validation_error", errResp.Error)
	assert.NotEmpty(t, errResp.Details)
}

func TestCreateMyProfile_ValidationError_MissingUsername(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{}
	handler := NewUserProfileHandler(mockUserProfile)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Missing username
	reqBody := CreateMyProfileRequest{
		Name:     "Test User",
		Username: "",
		IconPath: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "validation_error", errResp.Error)
	assert.NotEmpty(t, errResp.Details)
}

func TestCreateMyProfile_Unauthorized_MissingToken(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{}
	handler := NewUserProfileHandler(mockUserProfile)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Request without Authorization header
	reqBody := CreateMyProfileRequest{
		Name:     "Test User",
		Username: "testuser",
		IconPath: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "unauthorized", errResp.Error)
}

func TestCreateMyProfile_Unauthorized_InvalidToken(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{}
	handler := NewUserProfileHandler(mockUserProfile)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Request with invalid token
	reqBody := CreateMyProfileRequest{
		Name:     "Test User",
		Username: "testuser",
		IconPath: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_token", errResp.Error)
}

func TestCreateMyProfile_InternalServerError(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{
		createUserProfileFunc: func(ctx context.Context, uid int64, name string, username string, iconPath *string) (*domain.UserProfile, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewUserProfileHandler(mockUserProfile)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Create request
	reqBody := CreateMyProfileRequest{
		Name:     "Test User",
		Username: "testuser",
		IconPath: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 500, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "internal_server_error", errResp.Error)
}

func TestCreateMyProfile_WithGETMethod(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{}
	handler := NewUserProfileHandler(mockUserProfile)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Try GET instead of POST
	req := httptest.NewRequest("GET", "/api/v1/users/me/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Should return 405 Method Not Allowed
	assert.Equal(t, 405, resp.StatusCode)
}
