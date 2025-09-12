package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/internal/db"
)

var DB *pgx.Conn
var Queries *db.Queries

func ConnectDatabase(cfg *config.Config) (*pgx.Conn, *db.Queries) {
	var dsn string
	
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
	} else {
		dsn = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.DatabaseUser, cfg.DatabasePass, cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)
	}

	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = conn
	Queries = db.New(conn)
	return conn, Queries
}

func GetDB() *pgx.Conn {
	return DB
}

func GetQueries() *db.Queries {
	return Queries
}