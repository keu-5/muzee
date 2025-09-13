package infrastructure

import (
	"context"
	"log"

	"github.com/keu-5/muzee/backend/ent"
	"go.uber.org/fx"
)

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