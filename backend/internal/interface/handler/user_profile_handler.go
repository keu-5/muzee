package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/helper"
	"github.com/keu-5/muzee/backend/internal/usecase"
)

type UserProfileHandler struct {
	userProfileUC usecase.UserProfileUsecase
	validate      *validator.Validate
}

func NewUserProfileHandler(userProfileUC usecase.UserProfileUsecase) *UserProfileHandler {
	return &UserProfileHandler{
		userProfileUC: userProfileUC,
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
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Username string `json:"username" validate:"required,min=1,max=50"`
	IconPath string `json:"icon_path" validate:"max=255"`
}

type CreateMyProfileResponse struct {
	Message     string              `json:"message"`
	UserProfile UserProfileResponse `json:"user_profile"`
}

// CreateMyProfile creates a user profile for the authenticated user
//
//	@Summary		Create user profile
//	@Description	Creates a user profile for the currently authenticated user. Requires authentication via Bearer token (Authorization header) or HttpOnly cookie (access_token).
//	@Tags			user-profiles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Security		CookieAuth
//	@Param			request	body		CreateMyProfileRequest	true	"User profile information"
//	@Success		201		{object}	CreateMyProfileResponse
//	@Failure		400		{object}	helper.ErrorResponse
//	@Failure		401		{object}	helper.ErrorResponse
//	@Failure		500		{object}	helper.ErrorResponse
//	@Router			/v1/me/profile [post]
func (h *UserProfileHandler) CreateMyProfile(c *fiber.Ctx) error {
	// 1. リクエストパース、バリデーション
	var req CreateMyProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "bad_request",
			"message": "無効なリクエストボディです",
		})
	}
	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(helper.BuildValidationErrorResponse(err))
	}

	ctx := c.Context()

	// 2. ミドルウェアでlocalsに設定されたuser_idを取得
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
	}

	// 3. ユーザープロフィール作成
	var iconPath *string
	if req.IconPath != "" {
		iconPath = &req.IconPath
	}
	profile, err := h.userProfileUC.CreateUserProfile(ctx, userID, req.Name, req.Username, iconPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました",
		})
	}

	// 4. レスポンス返却
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
