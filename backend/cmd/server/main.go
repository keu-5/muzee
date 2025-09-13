package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/db"
	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()

	db.AutoMigrate(cfg)

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Listen(":8080")
}
