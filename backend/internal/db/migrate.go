package db

import (
	"context"
	"log"

	"entgo.io/ent/dialect"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/ent"
	"go.uber.org/fx"
)

func NewClient(cfg *config.Config) *ent.Client {
	client, err := ent.Open(dialect.Postgres, cfg.DSN())
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	return client
}

func AutoMigrate(lc fx.Lifecycle, client *ent.Client) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Schema.Create(ctx); err != nil {
				log.Fatalf("failed creating schema resources: %v", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})
}
