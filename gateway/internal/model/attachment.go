package model

import (
	"time"

	"github.com/google/uuid"
)

type ExtractionStatus string

const (
	ExtractionPending     ExtractionStatus = "pending"
	ExtractionQueued      ExtractionStatus = "queued"
	ExtractionDone        ExtractionStatus = "done"
	ExtractionFailed      ExtractionStatus = "failed"
	ExtractionUnsupported ExtractionStatus = "unsupported"
)

func (s ExtractionStatus) Valid() bool {
	switch s {
	case ExtractionPending, ExtractionQueued, ExtractionDone, ExtractionFailed, ExtractionUnsupported:
		return true
	}
	return false
}

type Attachment struct {
	ID               uuid.UUID        `db:"id"`
	UserID           uuid.UUID        `db:"user_id"`
	ConversationID   *uuid.UUID       `db:"conversation_id"`
	Filename         string           `db:"filename"`
	MIME             string           `db:"mime"`
	SizeBytes        int64            `db:"size_bytes"`
	ObjectKey        string           `db:"object_key"`
	TextKey          *string          `db:"text_key"`
	ExtractionStatus ExtractionStatus `db:"extraction_status"`
	ExtractionError  *string          `db:"extraction_error"`
	CreatedAt        time.Time        `db:"created_at"`
}
