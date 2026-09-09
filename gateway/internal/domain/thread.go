package domain

import (
	"unicode/utf8"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

type Thread struct {
	Messages []model.Message
}

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
