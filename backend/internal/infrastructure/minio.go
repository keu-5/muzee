package infrastructure

import (
	"context"
	"time"

	"github.com/keu-5/muzee/backend/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
)

func NewMinioClient(lc fx.Lifecycle, cfg *config.Config, logger *Logger) *minio.Client {
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
	})
	if err != nil {
		logger.Fatal("Failed to create MinIO client: ", err)
		return nil
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// 簡易Ping（Bucket一覧を取得して接続確認）
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			_, err := client.ListBuckets(ctx)
			if err != nil {
				logger.Fatal("Failed to connect to MinIO server: ", err)
				return err
			}

			logger.Info("Connected to MinIO server successfully")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing MinIO client")

			return nil
		},
	})

	return client
}
