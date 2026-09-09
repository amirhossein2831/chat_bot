package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

const (
	sessionPrefix = "session:"
	tokenBytes    = 32
)

type SessionRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewSessionRepository(client *redis.Client, ttl time.Duration) *SessionRepository {
	return &SessionRepository{client: client, ttl: ttl}
}

func (r *SessionRepository) Create(ctx context.Context, userID uuid.UUID) (string, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	token := base64.RawURLEncoding.EncodeToString(raw)
	if err := r.client.Set(ctx, key(token), userID.String(), r.ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (r *SessionRepository) UserID(ctx context.Context, token string) (uuid.UUID, error) {
	value, err := r.client.Get(ctx, key(token)).Result()
	if err != nil {
		return uuid.Nil, constants.ErrUnauthorized
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, constants.ErrUnauthorized
	}
	return id, nil
}

func (r *SessionRepository) Delete(ctx context.Context, token string) error {
	return r.client.Del(ctx, key(token)).Err()
}

func key(token string) string { return sessionPrefix + token }
