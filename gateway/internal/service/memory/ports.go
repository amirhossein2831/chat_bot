package memory

import (
	"context"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

type Repository interface {
	Instructions(ctx context.Context, userID uuid.UUID) (string, error)
	SetInstructions(ctx context.Context, userID uuid.UUID, instructions string) error
	Facts(ctx context.Context, userID uuid.UUID, limit int) ([]model.Fact, error)
	Add(ctx context.Context, fact model.Fact) error
	Delete(ctx context.Context, id, userID uuid.UUID) error
}

type Completer interface {
	Stream(ctx context.Context, req provider.ChatRequest) (<-chan provider.Delta, error)
}
