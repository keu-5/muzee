package infrastructure

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keu-5/muzee/backend/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
)

type StorageService struct {
	client *minio.Client
	cfg    *config.Config
	logger *Logger
}

func NewMinioClient(lc fx.Lifecycle, cfg *config.Config, logger *Logger) *minio.Client {
	// Remove http:// or https:// scheme from endpoint if present
	endpoint := strings.TrimPrefix(cfg.S3Endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	client, err := minio.New(endpoint, &minio.Options{
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

func NewStorageService(lc fx.Lifecycle, client *minio.Client, cfg *config.Config, logger *Logger) *StorageService {
	service := &StorageService{
		client: client,
		cfg:    cfg,
		logger: logger,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Initialize public and private buckets
			bucketsToCreate := []string{cfg.S3PublicBucket, cfg.S3PrivateBucket}
			for _, bucket := range bucketsToCreate {
				if err := service.EnsureBucket(ctx, bucket); err != nil {
					logger.Fatal(fmt.Sprintf("Failed to ensure bucket %s: ", bucket), err)
					return err
				}
			}
			logger.Info("Storage buckets initialized successfully")
			return nil
		},
	})

	return service
}

// EnsureBucket creates a bucket if it doesn't exist
func (s *StorageService) EnsureBucket(ctx context.Context, bucketName string) error {
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket %s: %w", bucketName, err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", bucketName, err)
		}
		s.logger.Info(fmt.Sprintf("Created bucket: %s", bucketName))
	}

	return nil
}

// UploadFile uploads a file to the specified bucket and returns the object path
func (s *StorageService) UploadFile(ctx context.Context, bucketName string, objectName string, file *multipart.FileHeader) error {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Upload to MinIO
	contentType := file.Header.Get("Content-Type")
	_, err = s.client.PutObject(
		ctx,
		bucketName,
		objectName,
		src,
		file.Size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Uploaded file: %s/%s", bucketName, objectName))
	return nil
}

// GetPresignedURL returns a presigned URL for accessing an object
func (s *StorageService) GetPresignedURL(ctx context.Context, bucketName string, objectName string, expiry time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(
		ctx,
		bucketName,
		objectName,
		expiry,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// DeleteFile deletes a file from the specified bucket
func (s *StorageService) DeleteFile(ctx context.Context, bucketName string, objectName string) error {
	err := s.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	s.logger.Info(fmt.Sprintf("Deleted file: %s/%s", bucketName, objectName))
	return nil
}

// GenerateUniqueObjectName generates a unique object name with the given prefix and extension
func (s *StorageService) GenerateUniqueObjectName(prefix string, filename string) string {
	ext := filepath.Ext(filename)
	return fmt.Sprintf("%s/%s%s", prefix, uuid.New().String(), ext)
}
