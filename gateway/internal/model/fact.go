package model

import (
	"time"

	"github.com/google/uuid"
)

type Fact struct {
	ID              uuid.UUID  `db:"id"`
	UserID          uuid.UUID  `db:"user_id"`
	Text            string     `db:"text"`
	TextNorm        string     `db:"text_norm"`
	SourceMessageID *uuid.UUID `db:"source_message_id"`
	Pinned          bool       `db:"pinned"`
	CreatedAt       time.Time  `db:"created_at"`
}
