package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/helper"
	"github.com/keu-5/muzee/backend/internal/usecase"
)

type UserHandler struct {
	userUC usecase.UserUsecase
}

func NewUserHandler(userUC usecase.UserUsecase) *UserHandler {
	return &UserHandler{
		userUC: userUC,
	}
}

type GetMeResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GetMe returns the current authenticated user's information
//
//	@Summary		Get current user
//	@Description	Returns the current authenticated user's information
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	GetMeResponse
//	@Failure		401	{object}	ErrorResponse
//	@Failure		500	{object}	ErrorResponse
//	@Router			/api/v1/users/me [get]
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	// ミドルウェアでlocalsに設定されたuser_idを取得
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ErrorResponse{
			Error:   "unauthorized",
			Message: "認証が必要です",
		})
	}

	ctx := c.Context()

	// ユーザー情報を取得
	user, err := h.userUC.GetUserByID(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(helper.ErrorResponse{
			Error:   "internal_server_error",
			Message: "サーバーエラーが発生しました",
		})
	}
	if user == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(helper.ErrorResponse{
			Error:   "user_not_found",
			Message: "ユーザーが見つかりません",
		})
	}

	// レスポンス返却
	return c.JSON(GetMeResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	})
}
