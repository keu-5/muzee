package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/db"
)

var Pool *pgxpool.Pool
var Queries *db.Queries

func ConnectDatabase(cfg *config.Config) (*pgxpool.Pool, error) {
	var dsn string
	
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
	} else {
		dsn = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.DatabaseUser, cfg.DatabasePass, cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal("Failed to parse database config:", err)
	}

	poolConfig.MaxConns = 30
	poolConfig.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	Pool = pool
	Queries = db.New(pool)
	return pool, nil
}

func GetPool() *pgxpool.Pool {
	return Pool
}

func GetQueries() *db.Queries {
	return Queries
}

func NewQueries(pool *pgxpool.Pool) *db.Queries {
	return db.New(pool)
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}