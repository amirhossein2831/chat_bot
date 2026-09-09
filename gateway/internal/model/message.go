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

type Usage struct {
	In            int
	Out           int
	CostUSDMicros int64
}

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

// Quota is settled from this transition, so a second one would double-debit.
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

func (m *Message) MarkFailed(reason string) {
	m.Status = StatusFailed
	m.Error = &reason
}
