// Package model holds the entities that have a table, together with the logic
// that belongs to each entity. It imports nothing else from this codebase.
package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

type Status string

const (
	StatusStreaming Status = "streaming"
	StatusComplete  Status = "complete"
	StatusFailed    Status = "failed"
)

// Usage is what a generation actually consumed. Cost is USD micro-cents as an
// integer: money never touches a float.
type Usage struct {
	In            int
	Out           int
	CostUSDMicros int64
}

// Message is one turn. Messages form a tree rather than a list: ParentID plus
// Conversation.HeadMessageID is what lets a regenerated answer become a
// sibling branch instead of overwriting history.
type Message struct {
	ID             uuid.UUID  `db:"id"`
	ConversationID uuid.UUID  `db:"conversation_id"`
	ParentID       *uuid.UUID `db:"parent_id"`
	Role           Role       `db:"role"`
	Content        string     `db:"content"`
	ContentNorm    string     `db:"content_norm"`
	Status         Status     `db:"status"`
	Model          string     `db:"model"`
	TokensIn       int        `db:"tokens_in"`
	TokensOut      int        `db:"tokens_out"`
	CostUSDMicros  int64      `db:"cost_usd_micros"`
	Error          *string    `db:"error"`
	CreatedAt      time.Time  `db:"created_at"`
}

func (m Message) IsTerminal() bool {
	return m.Status == StatusComplete || m.Status == StatusFailed
}

// MarkComplete finalises a streaming message. Completing an already-terminal
// message is rejected: it would mean a second settlement against the same
// generation, and quota is debited from this transition.
func (m *Message) MarkComplete(u Usage) error {
	if m.Status != StatusStreaming {
		return fmt.Errorf("%w: cannot complete a %s message", ErrInvalidTransition, m.Status)
	}
	m.Status = StatusComplete
	m.TokensIn = u.In
	m.TokensOut = u.Out
	m.CostUSDMicros = u.CostUSDMicros
	return nil
}

// MarkFailed is deliberately not guarded: a generation can fail from any
// state, including one already believed terminal, and recording that is more
// useful than refusing it.
func (m *Message) MarkFailed(reason string) {
	m.Status = StatusFailed
	m.Error = &reason
}
