package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret      string
	DatabaseURL    string
	DatabaseHost   string
	DatabasePort   string
	DatabaseName   string
	DatabaseUser   string
	DatabasePass   string
}

func LoadConfig() *Config {
	_ = godotenv.Load(".env")

	cfg := &Config{
		JWTSecret:    os.Getenv("JWT_SECRET"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		DatabaseHost: os.Getenv("DB_HOST"),
		DatabasePort: os.Getenv("DB_PORT"),
		DatabaseName: os.Getenv("DB_NAME"),
		DatabaseUser: os.Getenv("DB_USER"),
		DatabasePass: os.Getenv("DB_PASSWORD"),
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is not set in environment")
	}

	// Set default database values if not provided
	if cfg.DatabaseHost == "" {
		cfg.DatabaseHost = "localhost"
	}
	if cfg.DatabasePort == "" {
		cfg.DatabasePort = "5432"
	}
	if cfg.DatabaseName == "" {
		cfg.DatabaseName = "muzee"
	}
	if cfg.DatabaseUser == "" {
		cfg.DatabaseUser = "postgres"
	}

	return cfg
}
