package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

type UsageRepository struct {
	db *pgxpool.Pool
}

func NewUsageRepository(db *pgxpool.Pool) *UsageRepository {
	return &UsageRepository{db: db}
}

// Record appends to the ledger. This table is the source of truth for cost;
// the redis counters are a cache that can be rebuilt from it.
func (r *UsageRepository) Record(ctx context.Context, userID, messageID uuid.UUID, modelID string, usage model.Usage) error {
	_, err := r.db.Exec(ctx,
		`insert into usage_events (id, user_id, message_id, model, tokens_in, tokens_out, cost_usd_micros)
		 values ($1, $2, $3, $4, $5, $6, $7)`,
		uuid.New(), userID, messageID, modelID, usage.In, usage.Out, usage.CostUSDMicros)
	return err
}

// SpentSince rebuilds a spend figure from the ledger, which is how the redis
// counter is repaired after a crash or an expired key.
func (r *UsageRepository) SpentSince(ctx context.Context, userID uuid.UUID, since time.Time) (int64, error) {
	var total *int64
	err := r.db.QueryRow(ctx,
		`select sum(cost_usd_micros) from usage_events
		  where user_id = $1 and created_at >= $2`, userID, since).Scan(&total)
	if err != nil {
		return 0, err
	}
	if total == nil {
		return 0, nil
	}
	return *total, nil
}

func (r *UsageRepository) ActiveSubscription(ctx context.Context, userID uuid.UUID) (string, *time.Time, error) {
	var planCode string
	var expiresAt *time.Time

	err := r.db.QueryRow(ctx,
		`select plan_code, expires_at from subscriptions
		  where user_id = $1 and (expires_at is null or expires_at > now())
		  order by expires_at desc nulls first
		  limit 1`, userID).Scan(&planCode, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, nil
	}
	return planCode, expiresAt, err
}
