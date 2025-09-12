package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/database"
	"github.com/keu-5/muzee/backend/internal/router"
	"github.com/keu-5/muzee/backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	logger.Init()
	cfg := config.LoadConfig()

	// Wire-generated dependency injection
	handler, err := InitializeApp(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize app", zap.Error(err))
	}
	defer database.Close()

	app := fiber.New()

	router.SetupRoutes(app, handler)

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Gracefully shutting down...")
		_ = app.Shutdown()
	}()

	// Ping database to ensure connection
	client := database.GetClient().Debug()

	if err := client.Schema.Create(context.Background()); err != nil {
		logger.Error("Failed to create schema", zap.Error(err))
	}

	logger.Info("Connected to database successfully")
	logger.Info("Server starting on port 8080")

	if err := app.Listen(":8080"); err != nil {
		logger.Error("Server is shutting down", zap.Error(err))
	}

	logger.Info("Running cleanup tasks...")
	logger.Info("Server was successful shutdown.")
}
