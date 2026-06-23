package repository

import (
	"bytes"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/mobilefarm/af/phone-observer/internal/config"
)

type MinIOStorage struct {
	client *minio.Client
	bucket string
}

func NewMinIOStorage(cfg config.Config) (*MinIOStorage, func(), error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, nil, err
	}
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, nil, err
		}
	}
	return &MinIOStorage{client: client, bucket: cfg.MinioBucket}, func() {}, nil
}

func (m *MinIOStorage) Upload(ctx context.Context, key string, data []byte) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, key, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: "image/png"})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", m.bucket, key), nil
}

func (m *MinIOStorage) Ping(ctx context.Context) error {
	_, err := m.client.ListBuckets(ctx)
	return err
}

type NoopStorage struct{}

func NewNoopStorage() *NoopStorage { return &NoopStorage{} }

func (n *NoopStorage) Upload(_ context.Context, key string, _ []byte) (string, error) {
	return "noop://" + key, nil
}

func (n *NoopStorage) Ping(_ context.Context) error { return nil }
