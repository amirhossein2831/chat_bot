package quota

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

type Counters interface {
	Reserve(ctx context.Context, userID uuid.UUID, holdID string, amount, ceiling int64) (bool, error)
	Settle(ctx context.Context, userID uuid.UUID, holdID string, actual int64) error
	Release(ctx context.Context, userID uuid.UUID, holdID string) error
	IncrementDaily(ctx context.Context, userID uuid.UUID) (int64, error)
	DailyCount(ctx context.Context, userID uuid.UUID) (int64, error)
}

type Ledger interface {
	Record(ctx context.Context, userID, messageID uuid.UUID, modelID string, usage model.Usage) error
	ActiveSubscription(ctx context.Context, userID uuid.UUID) (string, *time.Time, error)
}
