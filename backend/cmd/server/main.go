package main

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/keu-5/muzee/backend/config"
	_ "github.com/keu-5/muzee/backend/docs"
	"github.com/keu-5/muzee/backend/internal/infrastructure"
	interfacepkg "github.com/keu-5/muzee/backend/internal/interface"
	"github.com/keu-5/muzee/backend/internal/interface/handler"
	"github.com/keu-5/muzee/backend/internal/repository"
	"github.com/keu-5/muzee/backend/internal/usecase"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

func NewFiberApp() *fiber.App {
	return fiber.New()
}

func LogConfigLoaded(cfg *config.Config, logger *infrastructure.Logger) {
	if cfg != nil {
		logger.Info("Configuration loaded successfully")
	}
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

func RegisterRoutes(
	app *fiber.App,
	testHandler *handler.TestHandler,
	authHandler *handler.AuthHandler,
	cfg *config.Config,
) {
	interfacepkg.RegisterRoutes(app, testHandler, authHandler, cfg)
}

// NewEmailSender provides EmailClient as EmailSender interface for fx
func NewEmailSender(emailClient *infrastructure.EmailClient) usecase.EmailSender {
	return emailClient
}

// @title						Muzee API
// @version					1.0
// @description				This is the API documentation for the Muzee application.
// @host						localhost
// @BasePath					/api
// @schemes					http
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
func main() {
	fx.New(
		fx.Provide(
			infrastructure.NewDevelopmentLogger,
			config.Load,
			infrastructure.NewClient,
			infrastructure.NewRedisClient,
			infrastructure.NewEmailClient,
			NewEmailSender, // EmailClient -> EmailSender interface adapter
			NewFiberApp,

			// Repository
			repository.NewTestRepository,
			repository.NewUserRepository,

			// Usecase
			usecase.NewTestUsecase,
			usecase.NewAuthUsecase,
			usecase.NewEmailUsecase,

			// Handler
			handler.NewTestHandler,
			handler.NewAuthHandler,
		),
		fx.Invoke(
			LogConfigLoaded,
			infrastructure.AutoMigrate,
			RegisterRoutes,
			StartServer,
		),
	).Run()
}