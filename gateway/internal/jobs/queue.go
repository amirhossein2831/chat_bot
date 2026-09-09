package jobs

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

type Queue struct {
	client *river.Client[pgx.Tx]
}

func NewQueue(client *river.Client[pgx.Tx]) *Queue {
	return &Queue{client: client}
}

func (q *Queue) Enqueue(ctx context.Context, attachmentID uuid.UUID) error {
	_, err := q.client.Insert(ctx, ExtractArgs{AttachmentID: attachmentID}, nil)
	return err
}
