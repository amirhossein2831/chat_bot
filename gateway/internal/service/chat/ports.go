package chat

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
	"github.com/amirhossein2831/chat_bot/gateway/internal/domain"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/quota"
)

type Conversations interface {
	Get(ctx context.Context, id, userID uuid.UUID) (model.Conversation, error)
	Thread(ctx context.Context, conversationID uuid.UUID, limit int) (domain.Thread, error)
	ThreadTo(ctx context.Context, messageID uuid.UUID, limit int) (domain.Thread, error)
	Message(ctx context.Context, id uuid.UUID) (model.Message, error)
	AppendPair(ctx context.Context, user, assistant model.Message) error
	AppendBranch(ctx context.Context, m model.Message) error
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

// Documents is the file service. Declared here because this is what needs it.
type Documents interface {
	Attach(ctx context.Context, ids []uuid.UUID, messageID, userID uuid.UUID) error
	TextFor(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) ([]string, error)
	TextOf(ctx context.Context, messageID uuid.UUID) ([]string, error)
}

type Recaller interface {
	Context(ctx context.Context, userID uuid.UUID) (domain.MemoryContext, error)
}

// Observer is told about finished exchanges so memory can learn from them. It
// is best effort and must never affect the conversation that triggered it.
type Metrics interface {
	ObserveFirstToken(model string, d time.Duration)
	ObserveStream(model, outcome string, d time.Duration)
	ObserveUsage(model string, in, out int, costMicros int64)
	IncUpstreamError(model string)
	IncQuotaRejection(reason string)
}

type Observer interface {
	Observe(userID, messageID uuid.UUID, userText, assistantText string)
}
