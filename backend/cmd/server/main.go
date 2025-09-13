package main

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/infrastructure"
	"github.com/keu-5/muzee/backend/internal/interface/handler"
	interfacepkg "github.com/keu-5/muzee/backend/internal/interface"
	"github.com/keu-5/muzee/backend/internal/repository"
	"github.com/keu-5/muzee/backend/internal/usecase"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

func NewFiberApp() *fiber.App {
	return fiber.New()
}

func StartServer(lc fx.Lifecycle, app *fiber.App, cfg *config.Config) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				app.Listen(":8080")
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown()
		},
	})
}

func RegisterRoutes(app *fiber.App, h *handler.TestHandler) {
	interfacepkg.RegisterRoutes(app, h)
}

func main() {
	fx.New(
		fx.Provide(
			config.Load,
			infrastructure.NewClient,
			NewFiberApp,
			repository.NewTestRepository,
			usecase.NewTestUsecase,
			handler.NewTestHandler,
		),
		fx.Invoke(
			infrastructure.AutoMigrate,
			RegisterRoutes,
			StartServer,
		),
	).Run()
}
