package db

import (
	"context"
	"log"

	"entgo.io/ent/dialect"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/ent"
)

func AutoMigrate(cfg *config.Config) {
	client, err := ent.Open(dialect.Postgres, cfg.DSN())
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
}
