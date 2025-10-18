package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/helper"
	"github.com/keu-5/muzee/backend/internal/util"
)

// AuthMiddleware verifies JWT tokens and sets user information in context
func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Authorizationヘッダーを取得
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(helper.ErrorResponse{
				Error:   "unauthorized",
				Message: "認証が必要です",
			})
		}

		// "Bearer "プレフィックスを削除
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(helper.ErrorResponse{
				Error:   "invalid_token_format",
				Message: "トークンの形式が正しくありません",
			})
		}

		// JWTを検証
		claims, err := util.ValidateAccessToken(tokenString, jwtSecret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(helper.ErrorResponse{
				Error:   "invalid_token",
				Message: "トークンが無効または期限切れです",
			})
		}

		// コンテキストにユーザー情報を設定
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)

		return c.Next()
	}
}
