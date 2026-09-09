package objectstore

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

type Store struct {
	client *minio.Client
	bucket string
	ready  atomic.Bool
}

func New(lc fx.Lifecycle, cfg *config.Config, log *zap.Logger) (*Store, error) {
	client, err := minio.New(cfg.Objects.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.Objects.AccessKey, cfg.Objects.SecretKey, ""),
		Secure: cfg.Objects.UseTLS,
	})
	if err != nil {
		return nil, err
	}

	store := &Store{client: client, bucket: cfg.Objects.Bucket}

	// Storage being down disables uploads, not the whole product: chat must
	// keep working when object storage is unreachable, as with the relay.
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := store.ensureBucket(ctx); err != nil {
				log.Warn("object storage unavailable, uploads disabled", zap.Error(err))
				return nil
			}
			store.ready.Store(true)
			return nil
		},
	})

	return store, nil
}

func (s *Store) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("checking bucket: %w", err)
	}
	if !exists {
		return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{})
	}
	return nil
}

// available retries once, so storage that comes up after the gateway starts is
// picked up without a restart.
func (s *Store) available(ctx context.Context) error {
	if s.ready.Load() {
		return nil
	}
	if err := s.ensureBucket(ctx); err != nil {
		return fmt.Errorf("%w: %s", constants.ErrStorageUnavailable, err)
	}
	s.ready.Store(true)
	return nil
}

// PresignPut returns a URL the browser uploads to directly, so file bytes
// never pass through the gateway.
func (s *Store) PresignPut(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if err := s.available(ctx); err != nil {
		return "", err
	}

	presigned, err := s.client.PresignedPutObject(ctx, s.bucket, key, ttl)
	if err != nil {
		return "", err
	}
	return presigned.String(), nil
}

func (s *Store) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if err := s.available(ctx); err != nil {
		return "", err
	}

	presigned, err := s.client.PresignedGetObject(ctx, s.bucket, key, ttl, url.Values{})
	if err != nil {
		return "", err
	}
	return presigned.String(), nil
}

func (s *Store) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := s.available(ctx); err != nil {
		return nil, err
	}

	return s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
}

func (s *Store) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if err := s.available(ctx); err != nil {
		return err
	}

	_, err := s.client.PutObject(ctx, s.bucket, key, body, size,
		minio.PutObjectOptions{ContentType: contentType})
	return err
}

func (s *Store) Stat(ctx context.Context, key string) (int64, error) {
	if err := s.available(ctx); err != nil {
		return 0, err
	}

	info, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}
