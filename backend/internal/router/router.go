package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/handler"
	"github.com/keu-5/muzee/backend/internal/middleware"
)

func SetupRoutes(app *fiber.App, h *handler.Handler) {
	app.Get("/health", handler.HealthCheck)

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// 認証関連
	auth := v1.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/register", h.Register)

	// 認証が必要なエンドポイント
	v1.Get("/tests", handler.TestAuth, middleware.JWTProtected())
}
