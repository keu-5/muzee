package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/database"
	"github.com/keu-5/muzee/backend/internal/router"
)

func main() {
	cfg := config.LoadConfig()

	// データベース接続を初期化
	pool, _ := database.ConnectDatabase(cfg)
	defer database.Close()

	app := fiber.New()

	router.SetupRoutes(app, cfg)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		_ = <-c
		log.Println("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Ping database to ensure connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to database successfully")
	log.Println("Server starting on port 8080")

	if err := app.Listen(":8080"); err != nil {
		log.Println("Server is shutting down:", err)
	}

	log.Println("Running cleanup tasks...")
	// Add cleanup tasks here if needed
	log.Println("Server was successful shutdown.")
}
