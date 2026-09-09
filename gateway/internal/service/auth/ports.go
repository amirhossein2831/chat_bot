package auth

import (
	"context"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

type Users interface {
	ByEmail(ctx context.Context, email string) (model.User, error)
	ByID(ctx context.Context, id uuid.UUID) (model.User, error)
	Create(ctx context.Context, user model.User, inviteCode string) error
}

type Sessions interface {
	Create(ctx context.Context, userID uuid.UUID) (string, error)
	UserID(ctx context.Context, token string) (uuid.UUID, error)
	Delete(ctx context.Context, token string) error
}
