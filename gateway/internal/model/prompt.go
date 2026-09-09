package model

import (
	"time"

	"github.com/google/uuid"
)

type Prompt struct {
	ID        uuid.UUID `db:"id"`
	Locale    string    `db:"locale"`
	Category  string    `db:"category"`
	Title     string    `db:"title"`
	Body      string    `db:"body"`
	SortOrder int       `db:"sort_order"`
	CreatedAt time.Time `db:"created_at"`
}
