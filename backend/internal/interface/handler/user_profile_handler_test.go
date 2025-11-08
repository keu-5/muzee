package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
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
	createUserProfileFunc   func(ctx context.Context, userID int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error)
	isUsernameAvailableFunc func(ctx context.Context, username string) (bool, error)
}

func (m *mockUserProfileUsecase) CreateUserProfile(ctx context.Context, userID int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error) {
	if m.createUserProfileFunc != nil {
		return m.createUserProfileFunc(ctx, userID, name, username, iconFile)
	}
	var iconPath *string
	if iconFile != nil {
		path := "profiles/" + iconFile.Filename
		iconPath = &path
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

func (m *mockUserProfileUsecase) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	if m.isUsernameAvailableFunc != nil {
		return m.isUsernameAvailableFunc(ctx, username)
	}
	return true, nil
}

func setupTestUserProfileApp(handler *UserProfileHandler, jwtSecret string) *fiber.App {
	app := fiber.New()
	app.Post("/api/v1/users/me/profile", middleware.AuthMiddleware(jwtSecret), handler.CreateMyProfile)
	app.Get("/api/v1/user-profiles/check-username", handler.CheckUsernameAvailability)
	return app
}

func TestNewUserProfileHandler(t *testing.T) {
	mockUserProfile := &mockUserProfileUsecase{}
	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.userProfileUC)
	assert.NotNil(t, handler.fileHelper)
}

func TestCreateMyProfile_Success(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{
		createUserProfileFunc: func(ctx context.Context, uid int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error) {
			return &domain.UserProfile{
				ID:        1,
				UserID:    uid,
				Name:      name,
				Username:  username,
				IconPath:  nil,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Create request with form data
	body := new(bytes.Buffer)
	body.WriteString("name=Test+User&username=testuser")
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
}

func TestCreateMyProfile_Success_WithoutIconPath(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{
		createUserProfileFunc: func(ctx context.Context, uid int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error) {
			return &domain.UserProfile{
				ID:        1,
				UserID:    uid,
				Name:      name,
				Username:  username,
				IconPath:  nil,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Create request without icon
	body := new(bytes.Buffer)
	body.WriteString("name=Test+User&username=testuser")
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
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
	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Missing name
	body := new(bytes.Buffer)
	body.WriteString("name=&username=testuser")
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Missing username
	body := new(bytes.Buffer)
	body.WriteString("name=Test+User&username=")
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Request without Authorization header
	body := new(bytes.Buffer)
	body.WriteString("name=Test+User&username=testuser")
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Request with invalid token
	body := new(bytes.Buffer)
	body.WriteString("name=Test+User&username=testuser")
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		createUserProfileFunc: func(ctx context.Context, uid int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error) {
			return nil, errors.New("database error")
		},
	}

	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Create request
	body := new(bytes.Buffer)
	body.WriteString("name=Test+User&username=testuser")
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
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

func TestCheckUsernameAvailability_Success_Available(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{
		isUsernameAvailableFunc: func(ctx context.Context, username string) (bool, error) {
			return true, nil
		},
	}

	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/user-profiles/check-username?username=testuser", nil)
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var response CheckUsernameAvailabilityResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.True(t, response.Available)
}

func TestCheckUsernameAvailability_Success_NotAvailable(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{
		isUsernameAvailableFunc: func(ctx context.Context, username string) (bool, error) {
			if username == "existinguser" {
				return false, nil
			}
			return true, nil
		},
	}

	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/user-profiles/check-username?username=existinguser", nil)
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var response CheckUsernameAvailabilityResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.False(t, response.Available)
}

func TestCheckUsernameAvailability_ValidationError_MissingUsername(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{}
	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Request without username query parameter
	req := httptest.NewRequest("GET", "/api/v1/user-profiles/check-username", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "validation_error", errResp.Error)
}

func TestCheckUsernameAvailability_ValidationError_EmptyUsername(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{}
	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Request with empty username
	req := httptest.NewRequest("GET", "/api/v1/user-profiles/check-username?username=", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "validation_error", errResp.Error)
}

func TestCheckUsernameAvailability_InternalServerError(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{
		isUsernameAvailableFunc: func(ctx context.Context, username string) (bool, error) {
			return false, errors.New("database error")
		},
	}

	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/user-profiles/check-username?username=testuser", nil)
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

func TestCheckUsernameAvailability_WithPOSTMethod(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{}
	mockFileHelper := helper.NewFileHelper("/tmp/test-uploads")
	handler := NewUserProfileHandler(mockUserProfile, mockFileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Try POST instead of GET
	req := httptest.NewRequest("POST", "/api/v1/user-profiles/check-username?username=testuser", nil)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Should return 405 Method Not Allowed
	assert.Equal(t, 405, resp.StatusCode)
}
