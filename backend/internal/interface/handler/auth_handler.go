package handler

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/usecase"
	"github.com/redis/go-redis/v9"
)

type SendCodeRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type SendCodeResponse struct {
	Message   string `json:"message"`
	Email     string `json:"email"`
	ExpiresIn int    `json:"expires_in"`
}

// TODO: globalにする
type ErrorResponse struct {
	Error   string                   `json:"error"`
	Message string                   `json:"message"`
	Details []map[string]interface{} `json:"details,omitempty"`
}

type AuthHandler struct {
	authUC      usecase.AuthUsecase
	emailUC     usecase.EmailUsecase
	redisClient *redis.Client
	validate    *validator.Validate
}

func NewAuthHandler(authUC usecase.AuthUsecase, emailUC usecase.EmailUsecase, redisClient *redis.Client) *AuthHandler {
	return &AuthHandler{
		authUC:      authUC,
		emailUC:     emailUC,
		redisClient: redisClient,
		validate:    validator.New(),
	}
}

// SendCode sends verification code to email
//
//	@Summary		Send verification code
//	@Description	Sends a 6-digit verification code to the email for signup
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		SendCodeRequest	true	"Email and password"
//	@Success		200		{object}	SendCodeResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		429		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/auth/signup/send-code [post]
func (h *AuthHandler) SendCode(c *fiber.Ctx) error {
	// 1. リクエストパース
	var req SendCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "invalid_request",
			Message: "リクエストの形式が正しくありません",
		})
	}

	// 2. バリデーション
	if err := h.validate.Struct(req); err != nil {
		details := make([]map[string]interface{}, 0)
		for _, err := range err.(validator.ValidationErrors) {
			details = append(details, map[string]interface{}{
				"field":   strings.ToLower(err.Field()),
				"message": getValidationMessage(err),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "validation_error",
			Message: "入力内容に誤りがあります",
			Details: details,
		})
	}

	ctx := c.Context()

	// 3. メールアドレスの重複チェック
	exists, err := h.authUC.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました。しばらく待ってから再度お試しください",
		})
	}
	if exists {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
			Error:   "email_already_exists",
			Message: "このメールアドレスは既に登録されています",
		})
	}

	// 4. レート制限チェック
	if err := h.checkRateLimit(ctx, req.Email); err != nil {
		return c.Status(fiber.StatusTooManyRequests).JSON(ErrorResponse{
			Error:   "rate_limit_exceeded",
			Message: "送信回数が多すぎます。しばらく待ってから再度お試しください",
		})
	}

	// 5. パスワードをハッシュ化
	passwordHash, err := h.authUC.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました",
		})
	}

	// 6. 6桁の確認コード生成
	code, err := generateVerificationCode()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました",
		})
	}

	// 7. Redisに保存（15分間）
	if err := h.saveSignupSession(ctx, req.Email, passwordHash, code); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました",
		})
	}

	// 8. メール送信
	if err := h.emailUC.SendVerificationCode(req.Email, code); err != nil {
		fmt.Printf("メール送信エラー: %v\n", err)
	}

	// 9. レスポンス返却
	return c.JSON(SendCodeResponse{
		Message:   "確認コードを送信しました。メールを確認してください。",
		Email:     req.Email,
		ExpiresIn: 900,
	})
}

// ヘルパーメソッド

type signupSessionData struct {
	PasswordHash string `json:"password_hash"`
	Code         string `json:"code"`
	CreatedAt    int64  `json:"created_at"`
}

func (h *AuthHandler) checkRateLimit(ctx context.Context, email string) error {
	rateLimitKey := fmt.Sprintf("rate_limit:send_code:%s", email)
	count, err := h.redisClient.Incr(ctx, rateLimitKey).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		h.redisClient.Expire(ctx, rateLimitKey, 5*time.Minute)
	}
	if count > 3 {
		return fmt.Errorf("rate limit exceeded")
	}
	return nil
}

func (h *AuthHandler) saveSignupSession(ctx context.Context, email, passwordHash, code string) error {
	sessionData := signupSessionData{
		PasswordHash: passwordHash,
		Code:         code,
		CreatedAt:    time.Now().Unix(),
	}
	dataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("signup:%s", email)
	return h.redisClient.Set(ctx, key, dataJSON, 15*time.Minute).Err()
}

// ユーティリティ関数

func generateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "必須項目です"
	case "email":
		return "有効なメールアドレスを入力してください"
	case "min":
		return fmt.Sprintf("%s文字以上で入力してください", fe.Param())
	case "max":
		return fmt.Sprintf("%s文字以内で入力してください", fe.Param())
	default:
		return "入力内容が正しくありません"
	}
}
