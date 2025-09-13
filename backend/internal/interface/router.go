package interfacepkg

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/internal/interface/handler"
)

func RegisterRoutes(app *fiber.App, h *handler.TestHandler) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	app.Post("/tests", h.Create)
	app.Get("/tests", h.GetAll)
}
