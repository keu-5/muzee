package interfacepkg

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/interface/handler"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func RegisterRoutes(
	app *fiber.App,
	testHandler *handler.TestHandler,
	authHandler *handler.AuthHandler,
	cfg *config.Config,
) {
	if cfg != nil && cfg.GOEnv == "development" {
		app.Get("/docs/*", fiberSwagger.WrapHandler)
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Post("/tests", testHandler.Create)
	app.Get("/tests", testHandler.GetAll)

	v1 := app.Group("/v1")
	auth := v1.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.RefreshToken)

	signup := auth.Group("/signup")
	signup.Post("/send-code", authHandler.SendCode)
	signup.Post("/verify-code", authHandler.VerifyCode)
}