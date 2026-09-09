package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

const attachmentColumns = `id, user_id, conversation_id, filename, mime, size_bytes,
	object_key, text_key, extraction_status, extraction_error, created_at`

type AttachmentRepository struct {
	db *pgxpool.Pool
}

func NewAttachmentRepository(db *pgxpool.Pool) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

func (r *AttachmentRepository) Create(ctx context.Context, a model.Attachment) error {
	_, err := r.db.Exec(ctx,
		`insert into attachments (id, user_id, conversation_id, filename, mime,
		                          size_bytes, object_key, extraction_status)
		 values ($1, $2, $3, $4, $5, $6, $7, $8)`,
		a.ID, a.UserID, a.ConversationID, a.Filename, a.MIME,
		a.SizeBytes, a.ObjectKey, a.ExtractionStatus)
	return err
}

func (r *AttachmentRepository) Get(ctx context.Context, id uuid.UUID) (model.Attachment, error) {
	rows, err := r.db.Query(ctx, `select `+attachmentColumns+` from attachments where id = $1`, id)
	if err != nil {
		return model.Attachment{}, err
	}

	attachment, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Attachment])
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Attachment{}, constants.ErrNotFound
	}
	return attachment, err
}

// GetOwned returns not-found rather than forbidden for another user's
// attachment, so the endpoint cannot be used to discover which ids exist.
func (r *AttachmentRepository) GetOwned(ctx context.Context, id, userID uuid.UUID) (model.Attachment, error) {
	rows, err := r.db.Query(ctx,
		`select `+attachmentColumns+` from attachments where id = $1 and user_id = $2`, id, userID)
	if err != nil {
		return model.Attachment{}, err
	}

	attachment, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[model.Attachment])
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Attachment{}, constants.ErrNotFound
	}
	return attachment, err
}

func (r *AttachmentRepository) SetStatus(ctx context.Context, id uuid.UUID, status model.ExtractionStatus, textKey *string, failure *string) error {
	_, err := r.db.Exec(ctx,
		`update attachments
		    set extraction_status = $2, text_key = coalesce($3, text_key), extraction_error = $4
		  where id = $1`,
		id, status, textKey, failure)
	return err
}

func (r *AttachmentRepository) CountToday(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`select count(*) from attachments
		  where user_id = $1 and created_at >= date_trunc('day', now())`, userID).Scan(&count)
	return count, err
}
