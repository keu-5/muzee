package handler

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/helper"
	"github.com/keu-5/muzee/backend/internal/usecase"
	"github.com/keu-5/muzee/backend/internal/util"
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

type AuthHandler struct {
	authUC        usecase.AuthUsecase
	userUC        usecase.UserUsecase
	emailUC       usecase.EmailUsecase
	sessionHelper *helper.SessionHelper
	validate      *validator.Validate
	jwtSecret     string
}

func NewAuthHandler(authUC usecase.AuthUsecase, userUC usecase.UserUsecase, emailUC usecase.EmailUsecase, sessionHelper *helper.SessionHelper, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authUC:        authUC,
		userUC:        userUC,
		emailUC:       emailUC,
		sessionHelper: sessionHelper,
		validate:      validator.New(),
		jwtSecret:     jwtSecret,
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
		return c.Status(fiber.StatusBadRequest).JSON(helper.ErrorResponse{
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
				"message": helper.GetValidationMessage(err),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(helper.ErrorResponse{
			Error:   "validation_error",
			Message: "入力内容に誤りがあります",
			Details: details,
		})
	}

	ctx := c.Context()

	// 3. メールアドレスの重複チェック
	exists, err := h.authUC.CheckEmailExists(ctx, req.Email)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました。しばらく待ってから再度お試しください",
		})
	}
	if exists {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ErrorResponse{
			Error:   "email_already_exists",
			Message: "このメールアドレスは既に登録されています",
		})
	}

	// 4. レート制限チェック
	if err := h.sessionHelper.CheckRateLimit(ctx, req.Email); err != nil {
		return c.Status(fiber.StatusTooManyRequests).JSON(helper.ErrorResponse{
			Error:   "rate_limit_exceeded",
			Message: "送信回数が多すぎます。しばらく待ってから再度お試しください",
		})
	}

	// 5. パスワードをハッシュ化
	passwordHash, err := h.authUC.HashPassword(req.Password)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました",
		})
	}

	// 6. 6桁の確認コード生成
	code, err := util.GenerateVerificationCode()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました",
		})
	}

	// 7. Redisに保存（15分間）
	if err := h.sessionHelper.SaveSignupSession(ctx, req.Email, passwordHash, code); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
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

type VerifyCodeRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
	Code  string `json:"code" validate:"required,len=6"`
}

type VerifyCodeResponse struct {
	Message      string          `json:"message"`
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	TokenType    string          `json:"token_type"`
	ExpiresIn    int             `json:"expires_in"`
	User         UserResponse    `json:"user"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// VerifyCode verifies the code and creates user account
//
//	@Summary		Verify code and create account
//	@Description	Verifies the 6-digit code and creates a user account, returning access and refresh tokens
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		VerifyCodeRequest	true	"Email and verification code"
//	@Success		201		{object}	VerifyCodeResponse
//	@Failure		400		{object}	ErrorResponse
//	@Failure		500		{object}	ErrorResponse
//	@Router			/api/v1/auth/signup/verify-code [post]
func (h *AuthHandler) VerifyCode(c *fiber.Ctx) error {
	// 1. リクエストパース
	var req VerifyCodeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ErrorResponse{
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
				"message": helper.GetValidationMessage(err),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(helper.ErrorResponse{
			Error:   "validation_error",
			Message: "入力内容に誤りがあります",
			Details: details,
		})
	}

	ctx := c.Context()

	// 3. Redisからサインアップセッションを取得
	sessionData, err := h.sessionHelper.GetSignupSession(ctx, req.Email)
	if err != nil || sessionData == nil {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ErrorResponse{
			Error:   "session_not_found",
			Message: "確認コードが無効または期限切れです。最初からやり直してください",
		})
	}

	// 4. コード照合
	if sessionData.Code != req.Code {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ErrorResponse{
			Error:   "invalid_code",
			Message: "確認コードが一致しません",
		})
	}

	// 5. ユーザー作成
	user, err := h.userUC.CreateUser(ctx, req.Email, sessionData.PasswordHash)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "アカウントの作成に失敗しました",
		})
	}

	// 6. JWT生成
	accessToken, err := util.GenerateAccessToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "トークンの生成に失敗しました",
		})
	}

	// 7. リフレッシュトークン生成
	refreshToken, err := util.GenerateRefreshToken()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "トークンの生成に失敗しました",
		})
	}

	// 8. Redisにリフレッシュトークンを保存（30日間）
	if err := h.sessionHelper.SaveRefreshToken(ctx, refreshToken, user.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "トークンの保存に失敗しました",
		})
	}

	// 9. サインアップセッションを削除
	if err := h.sessionHelper.DeleteSignupSession(ctx, req.Email); err != nil {
		// ログには記録するが、ユーザーにはエラーを返さない
		fmt.Printf("サインアップセッション削除エラー: %v\n", err)
	}

	// 10. レスポンス返却
	return c.Status(fiber.StatusCreated).JSON(VerifyCodeResponse{
		Message:      "アカウントが作成されました",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
		User: UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	})
}
