package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/util/textnorm"
)

const factColumns = `id, user_id, text, text_norm, source_message_id, pinned, created_at`

type MemoryRepository struct {
	db *pgxpool.Pool
}

func NewMemoryRepository(db *pgxpool.Pool) *MemoryRepository {
	return &MemoryRepository{db: db}
}

func (r *MemoryRepository) Instructions(ctx context.Context, userID uuid.UUID) (string, error) {
	var instructions string
	err := r.db.QueryRow(ctx,
		`select custom_instructions from user_settings where user_id = $1`, userID).
		Scan(&instructions)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return instructions, err
}

func (r *MemoryRepository) SetInstructions(ctx context.Context, userID uuid.UUID, instructions string) error {
	_, err := r.db.Exec(ctx,
		`insert into user_settings (user_id, custom_instructions, updated_at)
		 values ($1, $2, now())
		 on conflict (user_id) do update
		 set custom_instructions = excluded.custom_instructions, updated_at = now()`,
		userID, instructions)
	return err
}

func (r *MemoryRepository) Facts(ctx context.Context, userID uuid.UUID, limit int) ([]model.Fact, error) {
	rows, err := r.db.Query(ctx,
		`select `+factColumns+` from memory_facts
		  where user_id = $1
		  order by pinned desc, created_at desc
		  limit $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[model.Fact])
}

// Add is deduplicated on the normalised text, so relearning the same fact in
// different wording collapses to one row instead of accumulating.
func (r *MemoryRepository) Add(ctx context.Context, fact model.Fact) error {
	_, err := r.db.Exec(ctx,
		`insert into memory_facts (id, user_id, text, text_norm, source_message_id, pinned)
		 values ($1, $2, $3, $4, $5, $6)
		 on conflict (user_id, text_norm) do nothing`,
		fact.ID, fact.UserID, fact.Text, textnorm.Normalize(fact.Text),
		fact.SourceMessageID, fact.Pinned)
	return err
}

func (r *MemoryRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`delete from memory_facts where id = $1 and user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return constants.ErrNotFound
	}
	return nil
}
