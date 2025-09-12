package database

import (
	"database/sql"
	"fmt"

	entsql "entgo.io/ent/dialect/sql"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/ent"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var EntClient *ent.Client

func ConnectDatabase(cfg *config.Config) (*ent.Client, error) {
	var dsn string
	
	if cfg.DatabaseURL != "" {
		dsn = cfg.DatabaseURL
	} else {
		dsn = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.DatabaseUser, cfg.DatabasePass, cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(30)
	db.SetMaxIdleConns(5)

	drv := entsql.OpenDB("postgres", db)
	client := ent.NewClient(ent.Driver(drv))

	EntClient = client
	return client, nil
}

func GetClient() *ent.Client {
	return EntClient
}

func NewClient() *ent.Client {
	return EntClient
}

func Close() {
	if EntClient != nil {
		EntClient.Close()
	}
}