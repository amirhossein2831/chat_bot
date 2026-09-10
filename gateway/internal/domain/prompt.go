package domain

import (
	"strings"
	"unicode/utf8"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

const (
	maxInstructionChars = 1000
	maxFactChars        = 3000

	baseSystemPrompt = "You are a helpful assistant. Answer in the language the user writes in."
)

// MemoryContext is what the user has told the assistant to remember. It is
// assembled per request and never stored in this shape.
type MemoryContext struct {
	Instructions string
	Facts        []string
}

// PromptContext is everything that will be sent for one generation. It holds
// model types rather than provider types so this package stays independent of
// any client.
type PromptContext struct {
	System    string
	Messages  []model.Message
	truncated bool
}

func BuildPrompt(
	thread Thread,
	memory MemoryContext,
	documents []string,
	input model.Message,
) PromptContext {
	var system strings.Builder
	system.WriteString(baseSystemPrompt)

	if instructions := truncate(memory.Instructions, maxInstructionChars); instructions != "" {
		system.WriteString("\n\nUser instructions:\n")
		system.WriteString(instructions)
	}
	if facts := truncate(strings.Join(memory.Facts, "\n"), maxFactChars); facts != "" {
		system.WriteString("\n\nKnown about the user:\n")
		system.WriteString(facts)
	}

	// Documents sit in the system section rather than in the turn, so that what
	// is stored as the user's message stays what they typed. Each arrives
	// already wrapped as data the model must not take instructions from.
	for _, document := range documents {
		system.WriteString("\n\n")
		system.WriteString(document)
	}

	messages := make([]model.Message, 0, len(thread.Messages)+1)
	messages = append(messages, thread.Messages...)
	messages = append(messages, input)

	return PromptContext{System: system.String(), Messages: messages}
}

// Fit drops the oldest turns until the estimate fits, and reports false when
// even the current input alone cannot. The last message is never dropped.
func (p PromptContext) Fit(window int) (PromptContext, bool) {
	if window <= 0 {
		return p, true
	}

	systemCost := estimateTokens(p.System)
	for len(p.Messages) > 1 && systemCost+p.TokenEstimate() > window {
		p.Messages = p.Messages[1:]
		p.truncated = true
	}

	if systemCost+p.TokenEstimate() > window {
		return p, false
	}
	return p, true
}

func (p PromptContext) Truncated() bool { return p.truncated }

func (p PromptContext) TokenEstimate() int {
	total := 0
	for _, m := range p.Messages {
		total += estimateTokens(m.Content) + 4
	}
	return total
}

// estimateTokens deliberately overestimates: it feeds quota reservation, and a
// reservation that runs short is worse than one that holds too much.
func estimateTokens(s string) int {
	return utf8.RuneCountInString(s)/2 + 1
}

func truncate(s string, limit int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return string(runes[:limit])
}
