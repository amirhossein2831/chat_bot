package chat

import (
	"context"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
	"github.com/amirhossein2831/chat_bot/gateway/internal/domain"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/quota"
)

type Conversations interface {
	Get(ctx context.Context, id, userID uuid.UUID) (model.Conversation, error)
	Thread(ctx context.Context, conversationID uuid.UUID, limit int) (domain.Thread, error)
	AppendPair(ctx context.Context, user, assistant model.Message) error
	AppendPartial(ctx context.Context, id uuid.UUID, content string) error
	Finalize(ctx context.Context, m model.Message) error
}

type Completer interface {
	Stream(ctx context.Context, req provider.ChatRequest) (<-chan provider.Delta, error)
}

type Catalogue interface {
	ContextWindow(modelID string) int
}

type Budget interface {
	Reserve(ctx context.Context, userID uuid.UUID, modelID string, promptTokens, maxTokens int) (quota.Reservation, error)
	Settle(ctx context.Context, r quota.Reservation, messageID uuid.UUID, usage provider.Usage) (model.Usage, error)
	Release(ctx context.Context, r quota.Reservation) error
}

// Recaller supplies per-user memory. It may be nil until the memory service
// exists, in which case prompts carry no memory.
type Recaller interface {
	Context(ctx context.Context, userID uuid.UUID) (domain.MemoryContext, error)
}
