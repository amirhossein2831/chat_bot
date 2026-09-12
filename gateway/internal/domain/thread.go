package domain

import (
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

type Thread struct {
	Messages []model.Message
	// Siblings maps a message on the path to every message sharing its parent,
	// itself included and in the order they were written.
	Siblings map[uuid.UUID][]uuid.UUID
}

// Sibling is one branch point: a message and the parent it hangs from.
type Sibling struct {
	ID       uuid.UUID
	ParentID *uuid.UUID
}

func ThreadFrom(messages []model.Message) Thread {
	ordered := make([]model.Message, len(messages))
	for i, m := range messages {
		ordered[len(messages)-1-i] = m
	}
	return Thread{Messages: ordered}
}

// WithSiblings indexes the branches beside this path by the message on it.
func (t Thread) WithSiblings(all []Sibling) Thread {
	byParent := make(map[uuid.UUID][]uuid.UUID, len(all))
	var roots []uuid.UUID

	for _, s := range all {
		if s.ParentID == nil {
			roots = append(roots, s.ID)
			continue
		}
		byParent[*s.ParentID] = append(byParent[*s.ParentID], s.ID)
	}

	t.Siblings = make(map[uuid.UUID][]uuid.UUID, len(t.Messages))
	for _, m := range t.Messages {
		if m.ParentID == nil {
			t.Siblings[m.ID] = roots
			continue
		}
		t.Siblings[m.ID] = byParent[*m.ParentID]
	}
	return t
}

func (t Thread) Len() int { return len(t.Messages) }

func (t Thread) Head() (model.Message, bool) {
	if len(t.Messages) == 0 {
		return model.Message{}, false
	}
	return t.Messages[len(t.Messages)-1], true
}

// WithoutHead is what a regeneration builds its prompt from: the newest turn
// becomes the input again rather than staying part of the history.
func (t Thread) WithoutHead() Thread {
	if len(t.Messages) == 0 {
		return t
	}
	return Thread{Messages: t.Messages[:len(t.Messages)-1]}
}

func (t Thread) TokenEstimate() int {
	total := 0
	for _, m := range t.Messages {
		total += utf8.RuneCountInString(m.Content)/2 + 4
	}
	return total
}
