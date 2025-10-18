package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	GOEnv      string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	ResendEmailDomain string
	ResendAPIKey      string

	JWTSecret string
}

func Load() *Config {
	viper.SetDefault("POSTGRES_HOST", "db")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_USER", "appuser")
	viper.SetDefault("POSTGRES_PASSWORD", "apppassword")
	viper.SetDefault("POSTGRES_DB", "appdb")
	viper.SetDefault("GO_ENV", "development")

	viper.SetDefault("REDIS_ADDR", "redis:6379")
	viper.SetDefault("REDIS_PASSWORD", "redispassword")
	viper.SetDefault("REDIS_DB", 0)

	viper.SetDefault("RESEND_EMAIL_DOMAIN", "")
	viper.SetDefault("RESEND_API_KEY", "")

	viper.SetDefault("JWT_SECRET", "")

	viper.AutomaticEnv()

	viper.SetConfigName(".env.dev")
	viper.SetConfigType("env")
	viper.AddConfigPath("../deploy")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		// Config file not found or error reading - using defaults
	}

	return &Config{
		DBHost:     viper.GetString("POSTGRES_HOST"),
		DBPort:     viper.GetString("POSTGRES_PORT"),
		DBUser:     viper.GetString("POSTGRES_USER"),
		DBPassword: viper.GetString("POSTGRES_PASSWORD"),
		DBName:     viper.GetString("POSTGRES_DB"),
		GOEnv:      viper.GetString("GO_ENV"),

		RedisAddr:     viper.GetString("REDIS_ADDR"),
		RedisPassword: viper.GetString("REDIS_PASSWORD"),
		RedisDB:       viper.GetInt("REDIS_DB"),

		ResendEmailDomain: viper.GetString("RESEND_EMAIL_DOMAIN"),
		ResendAPIKey:      viper.GetString("RESEND_API_KEY"),

		JWTSecret: viper.GetString("JWT_SECRET"),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}
