//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/database"
	"github.com/keu-5/muzee/backend/internal/handler"
	"github.com/keu-5/muzee/backend/internal/repository"
	"github.com/keu-5/muzee/backend/internal/service"
)

// InitializeApp initializes the entire application with all dependencies
func InitializeApp(cfg *config.Config) (*handler.Handler, error) {
	wire.Build(
		database.ProviderSet,
		repository.ProviderSet,
		service.ProviderSet,
		handler.ProviderSet,
	)
	return nil, nil
}