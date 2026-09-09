package jobs

import (
	"context"

	"github.com/google/uuid"
)

type Queue struct {
	client InsertClient
}

func NewQueue(client InsertClient) *Queue {
	return &Queue{client: client}
}

func (q *Queue) Enqueue(ctx context.Context, attachmentID uuid.UUID) error {
	_, err := q.client.Insert(ctx, ExtractArgs{AttachmentID: attachmentID}, nil)
	return err
}
