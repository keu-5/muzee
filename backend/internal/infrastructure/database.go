package infrastructure

import (
	"log"

	"entgo.io/ent/dialect"
	"github.com/keu-5/muzee/backend/config"
	"github.com/keu-5/muzee/backend/ent"
)

func NewClient(cfg *config.Config) *ent.Client {
	client, err := ent.Open(dialect.Postgres, cfg.DSN())
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	return client
}