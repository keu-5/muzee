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
	createUserProfileFunc      func(ctx context.Context, userID int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error)
	getUserProfileByUserIDFunc func(ctx context.Context, userID int64) (*domain.UserProfile, error)
	isUsernameAvailableFunc    func(ctx context.Context, username string) (bool, error)
}

func (m *mockUserProfileUsecase) CreateUserProfile(ctx context.Context, userID int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error) {
	if m.createUserProfileFunc != nil {
		return m.createUserProfileFunc(ctx, userID, name, username, iconFile)
	}
	// Default: no icon file provided
	var iconPath *string
	if iconFile != nil {
		path := "user-icons/user_" + string(rune(userID)) + "/test-icon.jpg"
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

func (m *mockUserProfileUsecase) GetUserProfileByUserID(ctx context.Context, userID int64) (*domain.UserProfile, error) {
	if m.getUserProfileByUserIDFunc != nil {
		return m.getUserProfileByUserIDFunc(ctx, userID)
	}
	return &domain.UserProfile{
		ID:        1,
		UserID:    userID,
		Name:      "Test User",
		Username:  "testuser",
		IconPath:  nil,
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
	app.Get("/api/v1/users/me/profile", middleware.AuthMiddleware(jwtSecret), handler.GetMyProfile)
	app.Get("/api/v1/user-profiles/check-username", handler.CheckUsernameAvailability)
	return app
}

func TestNewUserProfileHandler(t *testing.T) {
	mockUserProfile := &mockUserProfileUsecase{}
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.userProfileUC)
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

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Create multipart form request
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Test User")
	writer.WriteField("username", "testuser")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Create multipart form request without icon
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Test User")
	writer.WriteField("username", "testuser")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Invalid form data
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", bytes.NewReader([]byte("invalid form data")))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "multipart/form-data")

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
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Missing name
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "")
	writer.WriteField("username", "testuser")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Missing username
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Test User")
	writer.WriteField("username", "")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Request without Authorization header
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Test User")
	writer.WriteField("username", "testuser")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Request with invalid token
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Test User")
	writer.WriteField("username", "testuser")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	req.Header.Set("Content-Type", writer.FormDataContentType())

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

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Create request
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Test User")
	writer.WriteField("username", "testuser")
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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

	mockUserProfile := &mockUserProfileUsecase{
		getUserProfileByUserIDFunc: func(ctx context.Context, uid int64) (*domain.UserProfile, error) {
			return &domain.UserProfile{
				ID:        1,
				UserID:    uid,
				Name:      "Test User",
				Username:  "testuser",
				IconPath:  nil,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Try GET instead of POST - should now call GetMyProfile endpoint
	req := httptest.NewRequest("GET", "/api/v1/users/me/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Should return 200 OK from GetMyProfile endpoint
	assert.Equal(t, 200, resp.StatusCode)
}

func TestCheckUsernameAvailability_Success_Available(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{
		isUsernameAvailableFunc: func(ctx context.Context, username string) (bool, error) {
			return true, nil
		},
	}

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
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

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
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
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
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
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
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

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
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
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Try POST instead of GET
	req := httptest.NewRequest("POST", "/api/v1/user-profiles/check-username?username=testuser", nil)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Should return 405 Method Not Allowed
	assert.Equal(t, 405, resp.StatusCode)
}

func TestCreateMyProfile_Success_WithIconFile(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	iconPath := "user-icons/user_123/test-icon.jpg"
	mockUserProfile := &mockUserProfileUsecase{
		createUserProfileFunc: func(ctx context.Context, uid int64, name string, username string, iconFile *multipart.FileHeader) (*domain.UserProfile, error) {
			// Verify that icon file was passed
			if iconFile == nil {
				t.Error("Expected iconFile to be non-nil")
			}
			return &domain.UserProfile{
				ID:        1,
				UserID:    uid,
				Name:      name,
				Username:  username,
				IconPath:  &iconPath,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Create multipart form request with icon
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Test User")
	writer.WriteField("username", "testuser")

	// Add a test image file with proper Content-Type
	partWriter, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="icon"; filename="test.jpg"`},
		"Content-Type":        {"image/jpeg"},
	})
	assert.NoError(t, err)
	// Write fake JPEG data (minimal valid JPEG header)
	fakeJPEG := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	partWriter.Write(fakeJPEG)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
	assert.Equal(t, iconPath, response.UserProfile.IconPath)
}

func TestCreateMyProfile_InvalidFileType(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{}
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Create multipart form request with invalid file type
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Test User")
	writer.WriteField("username", "testuser")

	// Add a test file with invalid extension
	part, err := writer.CreateFormFile("icon", "test.txt")
	assert.NoError(t, err)
	part.Write([]byte("This is not an image"))
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_file", errResp.Error)
}

func TestCreateMyProfile_FileSizeTooLarge(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{}
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)

	// Create custom app with larger body limit for testing
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10MB for test
	})
	app.Post("/api/v1/users/me/profile", middleware.AuthMiddleware(jwtSecret), handler.CreateMyProfile)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Create multipart form request with large file
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.WriteField("name", "Test User")
	writer.WriteField("username", "testuser")

	// Add a file that exceeds 5MB limit with proper Content-Type
	partWriter, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="icon"; filename="test.jpg"`},
		"Content-Type":        {"image/jpeg"},
	})
	assert.NoError(t, err)
	// Write JPEG header + large data (6MB)
	fakeJPEG := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	partWriter.Write(fakeJPEG)
	largeData := make([]byte, 6*1024*1024) // 6MB
	partWriter.Write(largeData)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()

		assert.Equal(t, 400, resp.StatusCode)

		var errResp helper.ErrorResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &errResp)
		assert.Equal(t, "invalid_file", errResp.Error)
	}
}

func TestGetMyProfile_Success(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	iconPath := "user-icons/user_123/icon.jpg"
	mockUserProfile := &mockUserProfileUsecase{
		getUserProfileByUserIDFunc: func(ctx context.Context, uid int64) (*domain.UserProfile, error) {
			return &domain.UserProfile{
				ID:        1,
				UserID:    uid,
				Name:      "Test User",
				Username:  "testuser",
				IconPath:  &iconPath,
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/users/me/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var response GetMyProfileResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.Contains(t, response.Message, "プロフィール")
	assert.Equal(t, int64(1), response.UserProfile.ID)
	assert.Equal(t, "Test User", response.UserProfile.Name)
	assert.Equal(t, "testuser", response.UserProfile.Username)
	assert.Equal(t, iconPath, response.UserProfile.IconPath)
}

func TestGetMyProfile_Success_WithoutIcon(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{
		getUserProfileByUserIDFunc: func(ctx context.Context, uid int64) (*domain.UserProfile, error) {
			return &domain.UserProfile{
				ID:        1,
				UserID:    uid,
				Name:      "Test User",
				Username:  "testuser",
				IconPath:  nil,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
	}

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/users/me/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var response GetMyProfileResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.Equal(t, "Test User", response.UserProfile.Name)
	assert.Equal(t, "testuser", response.UserProfile.Username)
	assert.Equal(t, "", response.UserProfile.IconPath)
}

func TestGetMyProfile_Unauthorized_MissingToken(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{}
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Request without Authorization header
	req := httptest.NewRequest("GET", "/api/v1/users/me/profile", nil)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "unauthorized", errResp.Error)
}

func TestGetMyProfile_Unauthorized_InvalidToken(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUserProfile := &mockUserProfileUsecase{}
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Request with invalid token
	req := httptest.NewRequest("GET", "/api/v1/users/me/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_token", errResp.Error)
}

func TestGetMyProfile_NotFound(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{
		getUserProfileByUserIDFunc: func(ctx context.Context, uid int64) (*domain.UserProfile, error) {
			return nil, nil
		},
	}

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/users/me/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "not_found", errResp.Error)
}

func TestGetMyProfile_InternalServerError(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{
		getUserProfileByUserIDFunc: func(ctx context.Context, uid int64) (*domain.UserProfile, error) {
			return nil, errors.New("database error")
		},
	}

	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/v1/users/me/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

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

func TestGetMyProfile_WithPOSTMethod(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUserProfile := &mockUserProfileUsecase{}
	fileHelper := helper.NewFileHelper()
	handler := NewUserProfileHandler(mockUserProfile, fileHelper)
	app := setupTestUserProfileApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := util.GenerateAccessToken(userID, email, false, jwtSecret)
	assert.NoError(t, err)

	// Try POST instead of GET
	req := httptest.NewRequest("POST", "/api/v1/users/me/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Should be handled by CreateMyProfile endpoint, so it should return 400 for bad request
	assert.NotEqual(t, 405, resp.StatusCode)
}
