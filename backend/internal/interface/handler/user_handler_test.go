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
	"github.com/golang-jwt/jwt/v5"
	"github.com/keu-5/muzee/backend/internal/domain"
	"github.com/keu-5/muzee/backend/internal/helper"
	"github.com/keu-5/muzee/backend/internal/interface/middleware"
	"github.com/keu-5/muzee/backend/internal/util"
	"github.com/stretchr/testify/assert"
)

func setupTestUserApp(handler *UserHandler, jwtSecret string) *fiber.App {
	app := fiber.New()
	app.Get("/api/v1/users/me", middleware.AuthMiddleware(jwtSecret), handler.GetMe)
	return app
}

func TestNewUserHandler(t *testing.T) {
	mockUser := &mockUserUsecase{}
	handler := NewUserHandler(mockUser)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.userUC)
}

func TestGetMe_Success(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUser := &mockUserUsecase{
		getUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
			return &domain.User{
				ID:           userID,
				Email:        email,
				PasswordHash: "hashed_password",
				CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, jwtSecret)

	// Create valid JWT token
	token, err := generateTestAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Create request with Authorization header
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var response GetMeResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.Equal(t, userID, response.ID)
	assert.Equal(t, email, response.Email)
	assert.Equal(t, "2024-01-01T00:00:00Z", response.CreatedAt)
	assert.Equal(t, "2024-01-02T00:00:00Z", response.UpdatedAt)
}

func TestGetMe_MissingAuthorizationHeader(t *testing.T) {
	jwtSecret := "test-secret-key"
	mockUser := &mockUserUsecase{}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, jwtSecret)

	// Create request without Authorization header
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "unauthorized", errResp.Error)
	assert.Equal(t, "認証が必要です", errResp.Message)
}

func TestGetMe_InvalidTokenFormat(t *testing.T) {
	jwtSecret := "test-secret-key"
	mockUser := &mockUserUsecase{}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, jwtSecret)

	// Create request with invalid token format (missing "Bearer " prefix)
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "InvalidTokenFormat")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_token_format", errResp.Error)
	assert.Equal(t, "トークンの形式が正しくありません", errResp.Message)
}

func TestGetMe_InvalidToken(t *testing.T) {
	jwtSecret := "test-secret-key"
	mockUser := &mockUserUsecase{}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, jwtSecret)

	// Create request with invalid token
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_token", errResp.Error)
	assert.Equal(t, "トークンが無効または期限切れです", errResp.Message)
}

func TestGetMe_ExpiredToken(t *testing.T) {
	jwtSecret := "test-secret-key"
	mockUser := &mockUserUsecase{}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, jwtSecret)

	// Create an expired token (with negative expiration)
	token, err := generateExpiredAccessToken(int64(123), "test@example.com", jwtSecret)
	assert.NoError(t, err)

	// Create request with expired token
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_token", errResp.Error)
}

func TestGetMe_WrongSecret(t *testing.T) {
	correctSecret := "correct-secret"
	wrongSecret := "wrong-secret"
	userID := int64(123)
	email := "test@example.com"

	mockUser := &mockUserUsecase{}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, correctSecret)

	// Create token with wrong secret
	token, err := generateTestAccessToken(userID, email, wrongSecret)
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_token", errResp.Error)
}

func TestGetMe_UserNotFound(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(999)
	email := "nonexistent@example.com"

	mockUser := &mockUserUsecase{
		getUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
			return nil, nil // User not found
		},
	}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, jwtSecret)

	// Create valid token
	token, err := generateTestAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "user_not_found", errResp.Error)
	assert.Equal(t, "ユーザーが見つかりません", errResp.Message)
}

func TestGetMe_DatabaseError(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	mockUser := &mockUserUsecase{
		getUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
			return nil, errors.New("database connection error")
		},
	}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, jwtSecret)

	// Create valid token
	token, err := generateTestAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
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
	assert.Equal(t, "サーバーエラーが発生しました", errResp.Message)
}

func TestGetMe_MultipleRequests(t *testing.T) {
	jwtSecret := "test-secret-key"
	userID := int64(123)
	email := "test@example.com"

	callCount := 0
	mockUser := &mockUserUsecase{
		getUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
			callCount++
			return &domain.User{
				ID:           userID,
				Email:        email,
				PasswordHash: "hashed_password",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, jwtSecret)

	// Create valid token
	token, err := generateTestAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Make multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
		resp.Body.Close()
	}

	// Verify that getUserByID was called 3 times
	assert.Equal(t, 3, callCount)
}

func TestGetMe_DifferentUsers(t *testing.T) {
	jwtSecret := "test-secret-key"

	mockUser := &mockUserUsecase{
		getUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
			return &domain.User{
				ID:           id,
				Email:        "user" + string(rune(id)) + "@example.com",
				PasswordHash: "hashed_password",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, jwtSecret)

	// Test with different user IDs
	userIDs := []int64{1, 2, 3}
	for _, userID := range userIDs {
		email := "user" + string(rune(userID)) + "@example.com"
		token, err := generateTestAccessToken(userID, email, jwtSecret)
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/api/v1/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)

		var response GetMeResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(bodyBytes, &response)
		assert.NoError(t, err)
		assert.Equal(t, userID, response.ID)
	}
}

func TestGetMe_WithPOSTMethod(t *testing.T) {
	jwtSecret := "test-secret-key"
	mockUser := &mockUserUsecase{}

	handler := NewUserHandler(mockUser)
	app := setupTestUserApp(handler, jwtSecret)

	userID := int64(123)
	email := "test@example.com"
	token, err := generateTestAccessToken(userID, email, jwtSecret)
	assert.NoError(t, err)

	// Try POST instead of GET
	req := httptest.NewRequest("POST", "/api/v1/users/me", bytes.NewReader([]byte("{}")))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Should return 405 Method Not Allowed
	assert.Equal(t, 405, resp.StatusCode)
}

// Helper function to generate a valid test access token
func generateTestAccessToken(userID int64, email string, secret string) (string, error) {
	// Import util package for token generation
	return generateAccessTokenHelper(userID, email, secret, 15*time.Minute)
}

// Helper function to generate an expired test access token
func generateExpiredAccessToken(userID int64, email string, secret string) (string, error) {
	// Generate token that expired 1 hour ago
	return generateAccessTokenHelper(userID, email, secret, -1*time.Hour)
}

// Helper to generate access token with custom expiration
func generateAccessTokenHelper(userID int64, email string, secret string, expiresIn time.Duration) (string, error) {
	// Use the util package to generate token
	// Since we can't easily override the expiration time in util.GenerateAccessToken,
	// we'll import the necessary packages and create our own token for testing

	// For now, we'll use the actual util function for valid tokens
	// and create a custom one for expired tokens
	if expiresIn > 0 {
		// Use the actual implementation from util package
		return generateValidToken(userID, email, secret)
	}

	// For expired tokens, create a custom implementation
	return generateCustomToken(userID, email, secret, expiresIn)
}

func generateValidToken(userID int64, email string, secret string) (string, error) {
	// Import from util package
	// This is a simplified version - in real tests we'd use util.GenerateAccessToken
	return generateCustomToken(userID, email, secret, 15*time.Minute)
}

func generateCustomToken(userID int64, email string, secret string, expiresIn time.Duration) (string, error) {
	// We need to import jwt package to create custom tokens
	// For testing purposes, we'll use the actual util.GenerateAccessToken
	// and accept that we can't easily test expired tokens without modifying the util package

	// For simplicity in this test, just use the util package
	// In a production environment, you might want to make the util function accept expiration as a parameter
	// or use a time.Now() function that can be mocked

	// Import the util package
	return generateTokenWithExpiration(userID, email, secret, expiresIn)
}

func generateTokenWithExpiration(userID int64, email string, secret string, expiresIn time.Duration) (string, error) {
	// Import jwt package
	// This requires importing "github.com/golang-jwt/jwt/v5"

	// For now, just use the util package implementation
	// In a real scenario, you'd want to refactor util.GenerateAccessToken to accept duration
	// or use dependency injection for time

	// Simplified: just import from util
	// Import: "github.com/keu-5/muzee/backend/internal/util"

	// Import jwt directly for test purposes
	return createJWTToken(userID, email, secret, expiresIn)
}

func createJWTToken(userID int64, email string, secret string, expiresIn time.Duration) (string, error) {
	// We'll need to import the jwt package and create a token manually
	// Import: "github.com/golang-jwt/jwt/v5"

	// For simplicity and to avoid circular imports or missing imports,
	// we'll just import from the util package directly
	// This is a test file, so we can use the util package

	// Let's just use the util package import
	// Import it at the top and use it

	// Since this is getting complex, let's simplify:
	// Just use util.GenerateAccessToken for valid tokens
	// For expired tokens, we can generate them separately

	// For now, import the util package at the top level
	// and use it in the helper

	// Actually, let's just import util at the package level
	// "github.com/keu-5/muzee/backend/internal/util"

	// And use it directly
	return createToken(userID, email, secret, expiresIn)
}

func createToken(userID int64, email string, secret string, expiresIn time.Duration) (string, error) {
	// Import jwt at package level: "github.com/golang-jwt/jwt/v5"
	// and create token with custom expiration

	// Since we need to add imports, let's just add them at the top
	// and implement this properly

	// For the test to work, we'll add the import and implementation
	// at the top of the file

	// For now, return a placeholder that we'll fix
	// when we add the proper imports

	// Let's just add the import and create the token properly
	// This will be added in the imports section

	// Import: github.com/golang-jwt/jwt/v5
	// Import: github.com/keu-5/muzee/backend/internal/util

	// Use util for simple case
	if expiresIn > 0 {
		// Just use util.GenerateAccessToken for valid tokens
		// For testing, we'll import util package
		return generateSimpleToken(userID, email, secret)
	}

	// For expired tokens, we need custom implementation
	// This requires jwt package which we'll import
	return generateExpiredToken(userID, email, secret, expiresIn)
}

// These will be implemented with proper imports
func generateSimpleToken(userID int64, email string, secret string) (string, error) {
	// This will use util.GenerateAccessToken
	// Import: "github.com/keu-5/muzee/backend/internal/util"
	return util.GenerateAccessToken(userID, email, false, secret)
}

func generateExpiredToken(userID int64, email string, secret string, expiresIn time.Duration) (string, error) {
	// This will use jwt package directly
	// Import: "github.com/golang-jwt/jwt/v5"

	// Create custom token with expiration
	expirationTime := time.Now().Add(expiresIn)
	claims := &util.JWTClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
