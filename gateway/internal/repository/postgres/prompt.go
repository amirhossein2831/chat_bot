package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

const promptColumns = `id, locale, category, title, body, sort_order, created_at`

type PromptRepository struct {
	db *pgxpool.Pool
}

func NewPromptRepository(db *pgxpool.Pool) *PromptRepository {
	return &PromptRepository{db: db}
}

func (r *PromptRepository) List(ctx context.Context, locale, category string) ([]model.Prompt, error) {
	rows, err := r.db.Query(ctx,
		`select `+promptColumns+`
		   from prompts
		  where locale = $1 and ($2 = '' or category = $2)
		  order by category, sort_order`, locale, category)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[model.Prompt])
}
