package domain

import (
	"unicode/utf8"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

// Thread is the path from a conversation's root to its head. It is assembled
// per request and never stored, which is why it lives here and not in model.
type Thread struct {
	Messages []model.Message
}

// ThreadFrom reverses the head-first ordering the recursive walk produces.
func ThreadFrom(messages []model.Message) Thread {
	ordered := make([]model.Message, len(messages))
	for i, m := range messages {
		ordered[len(messages)-1-i] = m
	}
	return Thread{Messages: ordered}
}

func (t Thread) Len() int { return len(t.Messages) }

func (t Thread) Head() (model.Message, bool) {
	if len(t.Messages) == 0 {
		return model.Message{}, false
	}
	return t.Messages[len(t.Messages)-1], true
}

// TokenEstimate deliberately overestimates: it feeds quota reservation, and a
// reservation that runs short is worse than one that holds too much.
func (t Thread) TokenEstimate() int {
	total := 0
	for _, m := range t.Messages {
		total += utf8.RuneCountInString(m.Content)/2 + 4
	}
	return total
}
