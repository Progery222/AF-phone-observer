package repository

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/mobilefarm/af/phone-observer/internal/config"
)

type MinIOStorage struct {
	client     *minio.Client
	bucket     string
	publicBase string
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
	if err := waitMinIOReady(ctx, client, cfg); err != nil {
		return nil, nil, err
	}
	return &MinIOStorage{
		client:     client,
		bucket:     cfg.MinioBucket,
		publicBase: strings.TrimRight(cfg.MinioPublicBaseURL, "/"),
	}, func() {}, nil
}

func waitMinIOReady(ctx context.Context, client *minio.Client, cfg config.Config) error {
	var lastErr error
	for attempt := 0; attempt < 40; attempt++ {
		if err := pingMinIOReady(cfg); err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		exists, err := client.BucketExists(ctx, cfg.MinioBucket)
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if !exists {
			if err := client.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{}); err != nil {
				lastErr = err
				time.Sleep(500 * time.Millisecond)
				continue
			}
			if err := client.SetBucketPolicy(ctx, cfg.MinioBucket, publicReadPolicy(cfg.MinioBucket)); err != nil {
				return err
			}
		}
		return nil
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("minio not ready")
}

func pingMinIOReady(cfg config.Config) error {
	scheme := "http"
	if cfg.MinioUseSSL {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s/minio/health/ready", scheme, cfg.MinioEndpoint)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("minio health ready %d", resp.StatusCode)
	}
	return nil
}

func publicReadPolicy(bucket string) string {
	return fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, bucket)
}

func (m *MinIOStorage) objectURL(key string) string {
	if m.publicBase != "" {
		return fmt.Sprintf("%s/%s/%s", m.publicBase, m.bucket, key)
	}
	return fmt.Sprintf("%s/%s", m.bucket, key)
}

func (m *MinIOStorage) Upload(ctx context.Context, key string, data []byte) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, key, bytes.NewReader(data), int64(len(data)),
		minio.PutObjectOptions{ContentType: "image/png"})
	if err != nil {
		return "", err
	}
	return m.objectURL(key), nil
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
