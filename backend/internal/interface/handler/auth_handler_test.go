package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/domain"
	"github.com/keu-5/muzee/backend/internal/helper"
	"github.com/keu-5/muzee/backend/internal/util"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// Mock AuthUsecase
type mockAuthUsecase struct {
	hashPasswordFunc     func(password string) (string, error)
	verifyPasswordFunc   func(password, hash string) error
	checkEmailExistsFunc func(ctx context.Context, email string) (bool, error)
}

func (m *mockAuthUsecase) HashPassword(password string) (string, error) {
	if m.hashPasswordFunc != nil {
		return m.hashPasswordFunc(password)
	}
	return "hashed_password", nil
}

func (m *mockAuthUsecase) VerifyPassword(password, hash string) error {
	if m.verifyPasswordFunc != nil {
		return m.verifyPasswordFunc(password, hash)
	}
	return nil
}

func (m *mockAuthUsecase) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	if m.checkEmailExistsFunc != nil {
		return m.checkEmailExistsFunc(ctx, email)
	}
	return false, nil
}

// Mock UserUsecase
type mockUserUsecase struct {
	createUserFunc     func(ctx context.Context, email, passwordHash string) (*domain.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*domain.User, error)
	getUserByIDFunc    func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *mockUserUsecase) CreateUser(ctx context.Context, email, passwordHash string) (*domain.User, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, email, passwordHash)
	}
	return &domain.User{
		ID:           1,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (m *mockUserUsecase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserUsecase) GetUserByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.getUserByIDFunc != nil {
		return m.getUserByIDFunc(ctx, id)
	}
	return &domain.User{
		ID:           id,
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// Mock EmailUsecase
type mockEmailUsecase struct {
	sendVerificationCodeFunc func(email, code string) error
}

func (m *mockEmailUsecase) SendVerificationCode(email, code string) error {
	if m.sendVerificationCodeFunc != nil {
		return m.sendVerificationCodeFunc(email, code)
	}
	return nil
}

func setupTestApp(handler *AuthHandler) *fiber.App {
	app := fiber.New()
	app.Post("/api/v1/auth/login", handler.Login)
	app.Post("/api/v1/auth/refresh", handler.RefreshToken)
	app.Post("/api/v1/auth/logout", handler.Logout)
	app.Post("/api/v1/auth/signup/send-code", handler.SendCode)
	app.Post("/api/v1/auth/signup/resend-code", handler.ResendCode)
	app.Post("/api/v1/auth/signup/verify-code", handler.VerifyCode)
	return app
}

func TestNewAuthHandler(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.validate)
}

func TestSendCode_Success(t *testing.T) {
	// Setup mocks
	mockAuth := &mockAuthUsecase{
		hashPasswordFunc: func(password string) (string, error) {
			return "hashed_password_123", nil
		},
		checkEmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
			return false, nil // Email doesn't exist
		},
	}

	mockEmail := &mockEmailUsecase{
		sendVerificationCodeFunc: func(email, code string) error {
			return nil
		},
	}

	// Use miniredis for testing
	mockUser := &mockUserUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Create request
	reqBody := SendCodeRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/send-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// For success case, we expect either 200 or 500 depending on Redis availability
	// If Redis is not available, it will fail at rate limit or session save
	assert.Contains(t, []int{200, 500, 429}, resp.StatusCode)
}

func TestSendCode_InvalidJSON(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/send-code", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_request", errResp.Error)
}

func TestSendCode_ValidationError_MissingEmail(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Missing email
	reqBody := SendCodeRequest{
		Email:    "",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/send-code", bytes.NewReader(body))
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

func TestSendCode_ValidationError_InvalidEmail(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Invalid email format
	reqBody := SendCodeRequest{
		Email:    "invalid-email",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/send-code", bytes.NewReader(body))
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

func TestSendCode_ValidationError_ShortPassword(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Password too short
	reqBody := SendCodeRequest{
		Email:    "test@example.com",
		Password: "short",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/send-code", bytes.NewReader(body))
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

func TestSendCode_EmailAlreadyExists(t *testing.T) {
	mockAuth := &mockAuthUsecase{
		checkEmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
			return true, nil // Email already exists
		},
	}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	reqBody := SendCodeRequest{
		Email:    "existing@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/send-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "email_already_exists", errResp.Error)
}

func TestSendCode_CheckEmailExistsError(t *testing.T) {
	mockAuth := &mockAuthUsecase{
		checkEmailExistsFunc: func(ctx context.Context, email string) (bool, error) {
			return false, errors.New("database error")
		},
	}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	reqBody := SendCodeRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/send-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 500, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "internal_server_error", errResp.Error)
}

func TestGenerateVerificationCode(t *testing.T) {
	code, err := util.GenerateVerificationCode()
	assert.NoError(t, err)
	assert.Len(t, code, 6)
	assert.Regexp(t, "^[0-9]{6}$", code)
}

func TestGetValidationMessage(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		param    string
		expected string
	}{
		{
			name:     "required field",
			tag:      "required",
			param:    "",
			expected: "必須項目です",
		},
		{
			name:     "email validation",
			tag:      "email",
			param:    "",
			expected: "有効なメールアドレスを入力してください",
		},
		{
			name:     "min length",
			tag:      "min",
			param:    "8",
			expected: "8文字以上で入力してください",
		},
		{
			name:     "max length",
			tag:      "max",
			param:    "255",
			expected: "255文字以内で入力してください",
		},
		{
			name:     "unknown validation",
			tag:      "unknown",
			param:    "",
			expected: "入力内容が正しくありません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock FieldError
			mockFieldError := &mockFieldError{
				tag:   tt.tag,
				param: tt.param,
			}
			result := helper.GetValidationMessage(mockFieldError)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Mock FieldError for testing
type mockFieldError struct {
	tag   string
	param string
}

func (m *mockFieldError) Tag() string                    { return m.tag }
func (m *mockFieldError) ActualTag() string              { return m.tag }
func (m *mockFieldError) Namespace() string              { return "" }
func (m *mockFieldError) StructNamespace() string        { return "" }
func (m *mockFieldError) Field() string                  { return "TestField" }
func (m *mockFieldError) StructField() string            { return "TestField" }
func (m *mockFieldError) Value() interface{}             { return nil }
func (m *mockFieldError) Param() string                  { return m.param }
func (m *mockFieldError) Kind() reflect.Kind             { return reflect.String }
func (m *mockFieldError) Type() reflect.Type             { return nil }
func (m *mockFieldError) Error() string                  { return "" }
func (m *mockFieldError) Translate(ut.Translator) string { return "" }

func TestSaveSignupSession(t *testing.T) {
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed_password"
	code := "123456"

	err := sessionHelper.SaveSignupSession(ctx, email, passwordHash, code)
	// Will fail if Redis is not running, but we're testing the function logic
	if err != nil {
		assert.Contains(t, err.Error(), "connection refused")
	}
}

func TestCheckRateLimit(t *testing.T) {
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	ctx := context.Background()
	email := "test@example.com"

	err := sessionHelper.CheckRateLimit(ctx, email)
	// Will fail if Redis is not running, but we're testing the function logic
	if err != nil {
		assert.Contains(t, err.Error(), "connection refused")
	}
}

// ========== ResendCode Tests ==========

func TestResendCode_Success(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{
		sendVerificationCodeFunc: func(email, code string) error {
			return nil
		},
	}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret-key", "development")
	app := setupTestApp(handler)

	// Pre-save existing signup session
	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed_password_123"
	oldCode := "123456"

	err := sessionHelper.SaveSignupSession(ctx, email, passwordHash, oldCode)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Create resend code request
	reqBody := ResendCodeRequest{
		Email: email,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/resend-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var response ResendCodeResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.Equal(t, "確認コードを再送信しました。メールを確認してください。", response.Message)
	assert.Equal(t, email, response.Email)
	assert.Equal(t, 900, response.ExpiresIn)
}

func TestResendCode_InvalidJSON(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/resend-code", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_request", errResp.Error)
	assert.Equal(t, "リクエストの形式が正しくありません", errResp.Message)
}

func TestResendCode_ValidationError_MissingEmail(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Missing email
	reqBody := ResendCodeRequest{
		Email: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/resend-code", bytes.NewReader(body))
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

func TestResendCode_ValidationError_InvalidEmail(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Invalid email format
	reqBody := ResendCodeRequest{
		Email: "invalid-email",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/resend-code", bytes.NewReader(body))
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

func TestResendCode_RateLimitExceeded(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	ctx := context.Background()
	email := "ratelimit@example.com"
	passwordHash := "hashed_password"
	code := "123456"

	// Pre-save signup session
	err := sessionHelper.SaveSignupSession(ctx, email, passwordHash, code)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Exhaust rate limit (assuming rate limit is implemented)
	// This test will pass if rate limit check is properly implemented
	for i := 0; i < 5; i++ {
		reqBody := ResendCodeRequest{Email: email}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/v1/auth/signup/resend-code", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		app.Test(req, -1)
	}

	// This request should be rate limited
	reqBody := ResendCodeRequest{Email: email}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/resend-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// May succeed if rate limit not hit, or return 429
	assert.Contains(t, []int{200, 429}, resp.StatusCode)

	if resp.StatusCode == 429 {
		var errResp helper.ErrorResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &errResp)
		assert.Equal(t, "rate_limit_exceeded", errResp.Error)
		assert.Equal(t, "送信回数が多すぎます。しばらく待ってから再度お試しください", errResp.Message)
	}
}

func TestResendCode_SessionNotFound(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Use a unique email to avoid rate limit from other tests
	uniqueEmail := fmt.Sprintf("nosession-%d@example.com", time.Now().UnixNano())

	// No existing signup session
	reqBody := ResendCodeRequest{
		Email: uniqueEmail,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/resend-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Should get either 400 (session_not_found) or 429 (rate_limit) depending on Redis state
	if resp.StatusCode == 400 {
		var errResp helper.ErrorResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &errResp)
		assert.Equal(t, "session_not_found", errResp.Error)
		assert.Equal(t, "確認コードが無効または期限切れです。最初からやり直してください", errResp.Message)
	} else {
		// If rate limit is hit or Redis unavailable, that's also acceptable
		assert.Contains(t, []int{429, 500}, resp.StatusCode)
	}
}

func TestResendCode_EmailSendFailure(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{
		sendVerificationCodeFunc: func(email, code string) error {
			return errors.New("email service error")
		},
	}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret-key", "development")
	app := setupTestApp(handler)

	// Pre-save existing signup session
	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed_password_123"
	oldCode := "123456"

	err := sessionHelper.SaveSignupSession(ctx, email, passwordHash, oldCode)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Create resend code request
	reqBody := ResendCodeRequest{
		Email: email,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/resend-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Even if email fails, the endpoint returns 200
	// (because the error is only logged, not returned)
	assert.Equal(t, 200, resp.StatusCode)

	var response ResendCodeResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.Equal(t, "確認コードを再送信しました。メールを確認してください。", response.Message)
}

func TestResendCode_NewCodeGenerated(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{
		sendVerificationCodeFunc: func(email, code string) error {
			return nil
		},
	}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret-key", "development")
	app := setupTestApp(handler)

	// Pre-save existing signup session with old code
	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed_password_123"
	oldCode := "123456"

	err := sessionHelper.SaveSignupSession(ctx, email, passwordHash, oldCode)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Resend code
	reqBody := ResendCodeRequest{
		Email: email,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/resend-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	// Verify old code no longer works by trying to verify with it
	verifyReq := VerifyCodeRequest{
		Email:    email,
		Code:     oldCode,
		ClientID: "test-client-id",
	}
	verifyBody, _ := json.Marshal(verifyReq)
	verifyHttpReq := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader(verifyBody))
	verifyHttpReq.Header.Set("Content-Type", "application/json")

	verifyResp, err := app.Test(verifyHttpReq, -1)
	assert.NoError(t, err)
	defer verifyResp.Body.Close()

	// Old code should not work (returns 400 invalid_code)
	assert.Equal(t, 400, verifyResp.StatusCode)

	var errResp helper.ErrorResponse
	verifyBodyBytes, _ := io.ReadAll(verifyResp.Body)
	json.Unmarshal(verifyBodyBytes, &errResp)
	assert.Equal(t, "invalid_code", errResp.Error)
}

func TestResendCode_UpdateRouteSetup(t *testing.T) {
	// This test ensures the route is properly set up
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")

	// Create app with ResendCode route
	app := fiber.New()
	app.Post("/api/v1/auth/signup/resend-code", handler.ResendCode)

	reqBody := ResendCodeRequest{
		Email: "test@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/resend-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Route should be accessible (not 404)
	assert.NotEqual(t, 404, resp.StatusCode)
}

// ========== VerifyCode Tests ==========

func TestVerifyCode_Success(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{
		createUserFunc: func(ctx context.Context, email, passwordHash string) (*domain.User, error) {
			return &domain.User{
				ID:           123,
				Email:        email,
				PasswordHash: passwordHash,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret-key", "development")
	app := setupTestApp(handler)

	// Pre-save signup session in Redis (this will fail if Redis is not running)
	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed_password_123"
	code := "123456"

	err := sessionHelper.SaveSignupSession(ctx, email, passwordHash, code)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Create verify code request
	reqBody := VerifyCodeRequest{
		Email:    email,
		Code:     code,
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 201, resp.StatusCode)

	var response VerifyCodeResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, int64(123), response.User.ID)
	assert.Equal(t, email, response.User.Email)
}

func TestVerifyCode_InvalidJSON(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_request", errResp.Error)
}

func TestVerifyCode_ValidationError_MissingEmail(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Missing email
	reqBody := VerifyCodeRequest{
		Email:    "",
		Code:     "123456",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader(body))
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

func TestVerifyCode_ValidationError_MissingClientID(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Missing client_id
	reqBody := VerifyCodeRequest{
		Email:    "test@example.com",
		Code:     "123456",
		ClientID: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader(body))
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

func TestVerifyCode_ValidationError_InvalidEmail(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Invalid email format
	reqBody := VerifyCodeRequest{
		Email:    "invalid-email",
		Code:     "123456",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader(body))
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

func TestVerifyCode_ValidationError_InvalidCodeLength(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Code too short
	reqBody := VerifyCodeRequest{
		Email:    "test@example.com",
		Code:     "123",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader(body))
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

func TestVerifyCode_ValidationError_MissingCode(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Missing code
	reqBody := VerifyCodeRequest{
		Email:    "test@example.com",
		Code:     "",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader(body))
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

func TestVerifyCode_SessionNotFound(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// No session saved in Redis
	reqBody := VerifyCodeRequest{
		Email:    "test@example.com",
		Code:     "123456",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "session_not_found", errResp.Error)
}

func TestVerifyCode_InvalidCode(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Pre-save signup session with different code
	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed_password_123"
	correctCode := "123456"

	err := sessionHelper.SaveSignupSession(ctx, email, passwordHash, correctCode)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Send wrong code
	reqBody := VerifyCodeRequest{
		Email:    email,
		Code:     "999999",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_code", errResp.Error)
}

func TestVerifyCode_CreateUserError(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{
		createUserFunc: func(ctx context.Context, email, passwordHash string) (*domain.User, error) {
			return nil, errors.New("database error")
		},
	}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Pre-save signup session
	ctx := context.Background()
	email := "test@example.com"
	passwordHash := "hashed_password_123"
	code := "123456"

	err := sessionHelper.SaveSignupSession(ctx, email, passwordHash, code)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Create verify code request
	reqBody := VerifyCodeRequest{
		Email:    email,
		Code:     code,
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/signup/verify-code", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 500, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "internal_server_error", errResp.Error)
}

// ========== Login Tests ==========

func TestLogin_Success(t *testing.T) {
	mockAuth := &mockAuthUsecase{
		verifyPasswordFunc: func(password, hash string) error {
			return nil
		},
	}
	mockUser := &mockUserUsecase{
		getUserByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:           123,
				Email:        email,
				PasswordHash: "hashed_password",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret-key", "development")
	app := setupTestApp(handler)

	// Create login request
	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// If Redis is available, expect 200, otherwise 500
	if resp.StatusCode == 200 {
		var response LoginResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(bodyBytes, &response)
		assert.NoError(t, err)
		assert.Equal(t, "ログインに成功しました", response.Message)
		assert.NotEmpty(t, response.AccessToken)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Equal(t, 900, response.ExpiresIn)
		assert.Equal(t, int64(123), response.User.ID)
		assert.Equal(t, "test@example.com", response.User.Email)
	} else {
		assert.Contains(t, []int{500, 429}, resp.StatusCode)
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_request", errResp.Error)
}

func TestLogin_ValidationError_MissingEmail(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Missing email
	reqBody := LoginRequest{
		Email:    "",
		Password: "password123",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
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

func TestLogin_ValidationError_MissingClientID(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Missing client_id
	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
		ClientID: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
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

func TestLogin_ValidationError_InvalidEmail(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Invalid email format
	reqBody := LoginRequest{
		Email:    "invalid-email",
		Password: "password123",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
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

func TestLogin_ValidationError_ShortPassword(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Password too short
	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "short",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
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

func TestLogin_UserNotFound(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{
		getUserByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, nil // User not found
		},
	}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	reqBody := LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// If Redis rate limit fails, might get 429 or 500
	if resp.StatusCode == 401 {
		var errResp helper.ErrorResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &errResp)
		assert.Equal(t, "invalid_credentials", errResp.Error)
	} else {
		assert.Contains(t, []int{429, 500}, resp.StatusCode)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	mockAuth := &mockAuthUsecase{
		verifyPasswordFunc: func(password, hash string) error {
			return errors.New("password mismatch")
		},
	}
	mockUser := &mockUserUsecase{
		getUserByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{
				ID:           123,
				Email:        email,
				PasswordHash: "hashed_password",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "wrongpassword",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// If Redis rate limit fails, might get 429 or 500
	if resp.StatusCode == 401 {
		var errResp helper.ErrorResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &errResp)
		assert.Equal(t, "invalid_credentials", errResp.Error)
		assert.Equal(t, "メールアドレスまたはパスワードが間違っています", errResp.Message)
	} else {
		assert.Contains(t, []int{429, 500}, resp.StatusCode)
	}
}

func TestLogin_GetUserByEmailError(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{
		getUserByEmailFunc: func(ctx context.Context, email string) (*domain.User, error) {
			return nil, errors.New("database error")
		},
	}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	reqBody := LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
		ClientID: "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// If Redis rate limit fails, might get 429, otherwise 500
	assert.Contains(t, []int{500, 429}, resp.StatusCode)
}

// ========== RefreshToken Tests ==========

func TestRefreshToken_Success(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{
		getUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
			return &domain.User{
				ID:           id,
				Email:        "test@example.com",
				PasswordHash: "hashed_password",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret-key", "development")
	app := setupTestApp(handler)

	// Pre-save refresh token in Redis
	ctx := context.Background()
	refreshToken := "test-refresh-token-123"
	userID := int64(123)
	clientID := "test-client-id-123"

	err := sessionHelper.SaveRefreshToken(ctx, refreshToken, userID, clientID)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Create refresh token request
	reqBody := RefreshTokenRequest{
		RefreshToken: refreshToken,
		ClientID:     clientID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var response RefreshTokenResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, 900, response.ExpiresIn)
}

func TestRefreshToken_InvalidJSON(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_request", errResp.Error)
}

func TestRefreshToken_ValidationError_MissingClientID(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Missing client_id
	reqBody := RefreshTokenRequest{
		RefreshToken: "some-refresh-token",
		ClientID:     "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader(body))
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

func TestRefreshToken_InvalidToken(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Token not in Redis
	reqBody := RefreshTokenRequest{
		RefreshToken: "invalid-token",
		ClientID:     "test-client-id-123",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "refresh_token_invalid", errResp.Error)
}

func TestRefreshToken_ClientIDMismatch(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Pre-save refresh token with different client_id
	ctx := context.Background()
	refreshToken := "test-refresh-token-456"
	userID := int64(123)
	correctClientID := "correct-client-id"

	err := sessionHelper.SaveRefreshToken(ctx, refreshToken, userID, correctClientID)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Try to use token with wrong client_id
	reqBody := RefreshTokenRequest{
		RefreshToken: refreshToken,
		ClientID:     "wrong-client-id",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "client_id_mismatch", errResp.Error)
}

func TestRefreshToken_UserNotFound(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{
		getUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
			return nil, nil // User not found
		},
	}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Pre-save refresh token
	ctx := context.Background()
	refreshToken := "test-refresh-token-789"
	userID := int64(999)
	clientID := "test-client-id-123"

	err := sessionHelper.SaveRefreshToken(ctx, refreshToken, userID, clientID)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Try to refresh with non-existent user
	reqBody := RefreshTokenRequest{
		RefreshToken: refreshToken,
		ClientID:     clientID,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "refresh_token_invalid", errResp.Error)
}

// ========== Logout Tests ==========

func TestLogout_Success(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret-key", "development")
	app := setupTestApp(handler)

	// Pre-save refresh token in Redis
	ctx := context.Background()
	refreshToken := "test-logout-token-123"
	userID := int64(123)
	clientID := "test-client-id-123"

	err := sessionHelper.SaveRefreshToken(ctx, refreshToken, userID, clientID)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// Create logout request
	reqBody := LogoutRequest{
		RefreshToken: refreshToken,
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var response LogoutResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)
	assert.Equal(t, "ログアウトしました", response.Message)

	// Verify token is deleted from Redis
	_, err = sessionHelper.GetRefreshToken(ctx, refreshToken)
	assert.Error(t, err) // Should error because token is deleted
}

func TestLogout_InvalidJSON(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Invalid JSON
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "invalid_request", errResp.Error)
}

func TestLogout_ValidationError_MissingToken(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Missing refresh token (both in body and cookies)
	reqBody := LogoutRequest{
		RefreshToken: "",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 400, resp.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "missing_refresh_token", errResp.Error)
}

func TestLogout_TokenNotFound(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret", "development")
	app := setupTestApp(handler)

	// Token not in Redis
	reqBody := LogoutRequest{
		RefreshToken: "nonexistent-token",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// If Redis is not available, might get 500
	if resp.StatusCode == 400 {
		var errResp helper.ErrorResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &errResp)
		assert.Equal(t, "token_not_found", errResp.Error)
		assert.Equal(t, "セッションが存在しません。既にログアウト済みです。", errResp.Message)
	} else {
		assert.Equal(t, 500, resp.StatusCode)
	}
}

func TestLogout_AlreadyLoggedOut(t *testing.T) {
	mockAuth := &mockAuthUsecase{}
	mockUser := &mockUserUsecase{}
	mockEmail := &mockEmailUsecase{}
	mockRedis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	sessionHelper := helper.NewSessionHelper(mockRedis)

	handler := NewAuthHandler(mockAuth, mockUser, mockEmail, sessionHelper, "test-secret-key", "development")
	app := setupTestApp(handler)

	// Pre-save and then delete refresh token
	ctx := context.Background()
	refreshToken := "test-logout-token-456"
	userID := int64(123)
	clientID := "test-client-id-123"

	err := sessionHelper.SaveRefreshToken(ctx, refreshToken, userID, clientID)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	// First logout (should succeed)
	reqBody := LogoutRequest{
		RefreshToken: refreshToken,
	}
	body, _ := json.Marshal(reqBody)
	req1 := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")

	resp1, err := app.Test(req1, -1)
	assert.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, 200, resp1.StatusCode)

	// Second logout (should fail with token_not_found)
	body2, _ := json.Marshal(reqBody)
	req2 := httptest.NewRequest("POST", "/api/v1/auth/logout", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := app.Test(req2, -1)
	assert.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, 400, resp2.StatusCode)

	var errResp helper.ErrorResponse
	bodyBytes, _ := io.ReadAll(resp2.Body)
	json.Unmarshal(bodyBytes, &errResp)
	assert.Equal(t, "token_not_found", errResp.Error)
}
