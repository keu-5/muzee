package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/database"
	"github.com/keu-5/muzee/backend/internal/router"
)

func main() {
	cfg := config.LoadConfig()

	// データベース接続を初期化
	_, _ = database.ConnectDatabase(cfg)

	app := fiber.New()

	router.SetupRoutes(app, cfg)

	log.Fatal(app.Listen(":8080"))
}
