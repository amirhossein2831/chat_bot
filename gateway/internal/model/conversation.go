package model

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID            uuid.UUID  `db:"id"`
	UserID        uuid.UUID  `db:"user_id"`
	Title         string     `db:"title"`
	TitleNorm     string     `db:"title_norm"`
	DefaultModel  string     `db:"default_model"`
	HeadMessageID *uuid.UUID `db:"head_message_id"`
	ArchivedAt    *time.Time `db:"archived_at"`
	CreatedAt     time.Time  `db:"created_at"`
}

func (c Conversation) IsArchived() bool { return c.ArchivedAt != nil }
