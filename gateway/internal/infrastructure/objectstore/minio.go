package objectstore

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
)

type Store struct {
	client *minio.Client
	bucket string
}

func New(lc fx.Lifecycle, cfg *config.Config) (*Store, error) {
	client, err := minio.New(cfg.Objects.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Objects.AccessKey, cfg.Objects.SecretKey, ""),
		Secure: cfg.Objects.UseTLS,
	})
	if err != nil {
		return nil, err
	}

	store := &Store{client: client, bucket: cfg.Objects.Bucket}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			exists, err := client.BucketExists(ctx, store.bucket)
			if err != nil {
				return fmt.Errorf("checking bucket: %w", err)
			}
			if !exists {
				return client.MakeBucket(ctx, store.bucket, minio.MakeBucketOptions{})
			}
			return nil
		},
	})

	return store, nil
}

// PresignPut returns a URL the browser uploads to directly, so file bytes
// never pass through the gateway.
func (s *Store) PresignPut(ctx context.Context, key string, ttl time.Duration) (string, error) {
	presigned, err := s.client.PresignedPutObject(ctx, s.bucket, key, ttl)
	if err != nil {
		return "", err
	}
	return presigned.String(), nil
}

func (s *Store) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	presigned, err := s.client.PresignedGetObject(ctx, s.bucket, key, ttl, url.Values{})
	if err != nil {
		return "", err
	}
	return presigned.String(), nil
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
}

func (s *Store) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucket, key, body, size,
		minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *Store) Stat(ctx context.Context, key string) (int64, error) {
	info, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}
