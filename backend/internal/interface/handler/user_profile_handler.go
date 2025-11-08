package handler

import (
	"mime/multipart"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/helper"
	"github.com/keu-5/muzee/backend/internal/usecase"
)

type UserProfileHandler struct {
	userProfileUC usecase.UserProfileUsecase
	fileHelper    *helper.FileHelper
	validate      *validator.Validate
}

func NewUserProfileHandler(userProfileUC usecase.UserProfileUsecase, fileHelper *helper.FileHelper) *UserProfileHandler {
	return &UserProfileHandler{
		userProfileUC: userProfileUC,
		fileHelper:    fileHelper,
		validate:      validator.New(),
	}
}

type UserProfileResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	IconPath string `json:"icon_path"`
}

type CreateMyProfileRequest struct {
	Name     string `form:"name" validate:"required,min=1,max=100"`
	Username string `form:"username" validate:"required,min=1,max=50"`
}

type CreateMyProfileResponse struct {
	Message     string              `json:"message"`
	UserProfile UserProfileResponse `json:"user_profile"`
}

// CreateMyProfile creates a user profile for the authenticated user
//
//	@Summary		Create user profile
//	@Description	Creates a user profile for the currently authenticated user. Requires authentication via Bearer token (Authorization header) or HttpOnly cookie (access_token). Accepts multipart form data with optional icon image file.
//	@Tags			user-profiles
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			name		formData	string	true	"User name (1-100 characters)"
//	@Param			username	formData	string	true	"Username (1-50 characters)"
//	@Param			icon		formData	file	false	"Profile icon image (max 5MB, JPEG/PNG/GIF/WebP)"
//	@Success		201			{object}	CreateMyProfileResponse
//	@Failure		400			{object}	helper.ErrorResponse
//	@Failure		401			{object}	helper.ErrorResponse
//	@Failure		500			{object}	helper.ErrorResponse
//	@Router			/v1/me/profile [post]
func (h *UserProfileHandler) CreateMyProfile(c *fiber.Ctx) error {
	ctx := c.Context()

	// 1. ミドルウェアでlocalsに設定されたuser_idを取得
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
	}

	// 2. フォームデータをパース
	var req CreateMyProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helper.ErrorResponse{
			Error:   "bad_request",
			Message: "無効なリクエストボディです",
		})
	}

	// 3. バリデーション
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helper.BuildValidationErrorResponse(err))
	}

	// 4. 画像ファイルの処理（オプショナル）
	var iconFile *multipart.FileHeader
	var err error
	iconFile, err = c.FormFile("icon")
	if err == nil {
		// ファイルが提供されている場合、バリデーション
		if err := h.fileHelper.ValidateImageFile(iconFile); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(helper.ErrorResponse{
				Error:   "invalid_file",
				Message: err.Error(),
			})
		}
	}

	// 5. ユーザープロフィール作成
	profile, err := h.userProfileUC.CreateUserProfile(ctx, userID, req.Name, req.Username, iconFile)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました",
		})
	}

	// 6. レスポンス返却
	iconPathStr := ""
	if profile.IconPath != nil {
		iconPathStr = *profile.IconPath
	}
	res := CreateMyProfileResponse{
		Message: "ユーザープロフィールが作成されました",
		UserProfile: UserProfileResponse{
			ID:       profile.ID,
			Name:     profile.Name,
			Username: profile.Username,
			IconPath: iconPathStr,
		},
	}
	return c.Status(fiber.StatusCreated).JSON(res)
}

type CheckUsernameAvailabilityRequest struct {
	Username string `query:"username" validate:"required,min=1,max=50"`
}

type CheckUsernameAvailabilityResponse struct {
	Available bool `json:"available"`
}

// CheckUsernameAvailability checks if a given username is available
//
//	@Summary		Check username availability
//	@Description	Checks whether the specified username is available for registration. This endpoint does not require authentication.
//	@Tags			user-profiles
//	@Accept			json
//	@Produce		json
//	@Param			username	query		string	true	"Username to check (1–50 characters)"
//	@Success		200			{object}	CheckUsernameAvailabilityResponse
//	@Failure		400			{object}	helper.ErrorResponse
//	@Failure		500			{object}	helper.ErrorResponse
//	@Router			/v1/user-profiles/check-username [get]
func (h *UserProfileHandler) CheckUsernameAvailability(c *fiber.Ctx) error {
	// 1. リクエストパース、バリデーション
	var req CheckUsernameAvailabilityRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "bad_request",
			"message": "無効なクエリパラメータです",
		})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helper.BuildValidationErrorResponse(err))
	}

	ctx := c.Context()

	// 2. ユーザーネームの利用可能性チェック
	available, err := h.userProfileUC.IsUsernameAvailable(ctx, req.Username)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました",
		})
	}

	// 3. レスポンス返却
	res := CheckUsernameAvailabilityResponse{
		Available: available,
	}
	return c.Status(fiber.StatusOK).JSON(res)
}
