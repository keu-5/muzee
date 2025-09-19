package interfacepkg

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/interface/handler"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

func RegisterRoutes(app *fiber.App, h *handler.TestHandler, cfg *config.Config) {
	if cfg != nil && cfg.GOEnv == "development" {
		app.Get("/docs/*", fiberSwagger.WrapHandler)
	}

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Post("/tests", h.Create)
	app.Get("/tests", h.GetAll)
    app.Get("/swagger-test", h.SwaggerTest)
}
