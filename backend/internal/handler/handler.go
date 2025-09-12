package handler

import (
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/service"
)

type Handler struct {
	authService *service.AuthService
	config      *config.Config
}

func NewHandler(authService *service.AuthService, config *config.Config) *Handler {
	return &Handler{
		authService: authService,
		config:      config,
	}
}