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
}

func Load() *Config {
	viper.SetDefault("POSTGRES_HOST", "db")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_USER", "appuser")
	viper.SetDefault("POSTGRES_PASSWORD", "apppassword")
	viper.SetDefault("POSTGRES_DB", "appdb")

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
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}
