package infrastructure

import (
	"entgo.io/ent/dialect"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/ent"
)

func NewClient(cfg *config.Config, logger *Logger) *ent.Client {
	client, err := ent.Open(dialect.Postgres, cfg.DSN())
	if err != nil {
		logger.Fatalf("failed opening connection to postgres: %v", err)
	}
	return client
}
