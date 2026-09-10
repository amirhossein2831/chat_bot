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

	charsPerToken = 2
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
	// Kept apart from the documents so that trimming to fit can cut the
	// documents without touching the instructions.
	instructions string
	documents    []string

	Messages  []model.Message
	truncated bool
}

// System is everything the model is told before the conversation itself.
func (p PromptContext) System() string {
	var out strings.Builder
	out.WriteString(p.instructions)

	for _, document := range p.documents {
		out.WriteString("\n\n")
		out.WriteString(document)
	}
	return out.String()
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

	messages := make([]model.Message, 0, len(thread.Messages)+1)
	messages = append(messages, thread.Messages...)
	messages = append(messages, input)

	// Documents sit in the system section rather than in the turn, so what is
	// stored as the user's message stays what they typed. Each arrives already
	// wrapped as data the model must not take instructions from.
	return PromptContext{
		instructions: system.String(),
		documents:    documents,
		Messages:     messages,
	}
}

// Fit reduces the prompt until it fits the model's window, in the order that
// costs the answer least: older turns first, then the documents.
//
// Cutting a document short and saying so beats refusing to answer. A thirty
// page contract does not fit any window, and a user who attached one wants an
// answer about what did fit, not an error.
func (p PromptContext) Fit(window int) (PromptContext, bool) {
	if window <= 0 {
		return p, true
	}

	fits := func() bool { return estimateTokens(p.System())+p.TokenEstimate() <= window }

	for len(p.Messages) > 1 && !fits() {
		p.Messages = p.Messages[1:]
		p.truncated = true
	}
	if fits() {
		return p, true
	}

	room := window - estimateTokens(p.instructions) - p.TokenEstimate()
	if room <= 0 {
		return p, false
	}

	p.documents = trimDocuments(p.documents, room)
	p.truncated = true
	return p, fits()
}

// trimDocuments fills the room available in order, so the first document
// attached is the one kept whole.
func trimDocuments(documents []string, room int) []string {
	const marker = "\n… (truncated)"

	budget := room * charsPerToken
	trimmed := make([]string, 0, len(documents))

	for _, document := range documents {
		if budget <= len(marker) {
			break
		}

		runes := []rune(document)
		if len(runes) <= budget {
			trimmed = append(trimmed, document)
			budget -= len(runes)
			continue
		}

		trimmed = append(trimmed, string(runes[:budget-len(marker)])+marker)
		budget = 0
	}
	return trimmed
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
	return utf8.RuneCountInString(s)/charsPerToken + 1
}

func truncate(s string, limit int) string {
	s = strings.TrimSpace(s)
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return string(runes[:limit])
}
