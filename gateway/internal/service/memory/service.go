package memory

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
	"github.com/amirhossein2831/chat_bot/gateway/internal/domain"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

const (
	// All facts are injected rather than retrieved. Vector retrieval would need
	// an embedding model whose Persian quality is the weakest link in the open
	// stack, and it would fail silently by recalling the wrong fact.
	maxInjectedFacts = 50

	maxInstructionLength = 1000
	maxFactLength        = 200
	maxFactsPerExchange  = 3

	extractionTimeout = 60 * time.Second

	extractionPrompt = `Extract at most three durable facts about the user from this exchange.
A fact is something still true next week: their job, language, preferences, ongoing projects.
Ignore anything transient, and ignore the assistant's own statements.
Reply with a JSON array of short strings and nothing else. Reply [] if there is nothing worth keeping.`
)

type Service struct {
	repo           Repository
	completer      Completer
	extractorModel string
	log            *zap.Logger
}

func NewService(repo Repository, completer Completer, extractorModel string, log *zap.Logger) *Service {
	return &Service{repo: repo, completer: completer, extractorModel: extractorModel, log: log}
}

// Context is what the chat service injects into every prompt.
func (s *Service) Context(ctx context.Context, userID uuid.UUID) (domain.MemoryContext, error) {
	instructions, err := s.repo.Instructions(ctx, userID)
	if err != nil {
		return domain.MemoryContext{}, err
	}

	facts, err := s.repo.Facts(ctx, userID, maxInjectedFacts)
	if err != nil {
		return domain.MemoryContext{}, err
	}

	texts := make([]string, 0, len(facts))
	for _, fact := range facts {
		texts = append(texts, fact.Text)
	}

	return domain.MemoryContext{Instructions: instructions, Facts: texts}, nil
}

func (s *Service) Instructions(ctx context.Context, userID uuid.UUID) (string, error) {
	return s.repo.Instructions(ctx, userID)
}

func (s *Service) SetInstructions(ctx context.Context, userID uuid.UUID, instructions string) error {
	return s.repo.SetInstructions(ctx, userID, truncate(instructions, maxInstructionLength))
}

func (s *Service) Facts(ctx context.Context, userID uuid.UUID) ([]model.Fact, error) {
	return s.repo.Facts(ctx, userID, maxInjectedFacts)
}

func (s *Service) Add(ctx context.Context, userID uuid.UUID, text string) (model.Fact, error) {
	fact := model.Fact{
		ID:     uuid.New(),
		UserID: userID,
		Text:   truncate(text, maxFactLength),
	}
	return fact, s.repo.Add(ctx, fact)
}

func (s *Service) Delete(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

// Observe runs the extractor over a finished exchange. It uses the local model
// so it never spends foreign currency, and it is best effort: a failure here
// must never affect the conversation that triggered it.
func (s *Service) Observe(userID, messageID uuid.UUID, userText, assistantText string) {
	if s.completer == nil || s.extractorModel == "" {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), extractionTimeout)
		defer cancel()

		facts, err := s.extract(ctx, userText, assistantText)
		if err != nil {
			s.log.Debug("fact extraction failed", zap.Error(err))
			return
		}

		for _, text := range facts {
			fact := model.Fact{
				ID:              uuid.New(),
				UserID:          userID,
				Text:            truncate(text, maxFactLength),
				SourceMessageID: &messageID,
			}
			if err := s.repo.Add(ctx, fact); err != nil {
				s.log.Debug("storing fact failed", zap.Error(err))
			}
		}
	}()
}

func (s *Service) extract(ctx context.Context, userText, assistantText string) ([]string, error) {
	deltas, err := s.completer.Stream(ctx, provider.ChatRequest{
		Model:  s.extractorModel,
		System: extractionPrompt,
		Messages: []provider.Message{{
			Role:    provider.RoleUser,
			Content: "User said: " + userText + "\n\nAssistant replied: " + assistantText,
		}},
		MaxTokens: 256,
	})
	if err != nil {
		return nil, err
	}

	var reply strings.Builder
	for delta := range deltas {
		if delta.Type == provider.DeltaText {
			reply.WriteString(delta.Text)
		}
	}

	var facts []string
	if err := json.Unmarshal([]byte(jsonArray(reply.String())), &facts); err != nil {
		return nil, err
	}
	if len(facts) > maxFactsPerExchange {
		facts = facts[:maxFactsPerExchange]
	}
	return facts, nil
}

// jsonArray pulls the array out of a reply that may be wrapped in prose or a
// code fence, which small models do regardless of instructions.
func jsonArray(reply string) string {
	start := strings.Index(reply, "[")
	end := strings.LastIndex(reply, "]")
	if start < 0 || end < start {
		return "[]"
	}
	return reply[start : end+1]
}

func truncate(s string, limit int) string {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) <= limit {
		return s
	}
	return string(runes[:limit])
}
