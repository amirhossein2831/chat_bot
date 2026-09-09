package chat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	"github.com/amirhossein2831/chat_bot/gateway/internal/domain"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/quota"
)

const (
	partialFlushInterval = 500 * time.Millisecond
	generationTimeout    = 10 * time.Minute
	threadLimit          = 200
	defaultMaxTokens     = 2048
)

type SendCommand struct {
	UserID         uuid.UUID
	ConversationID uuid.UUID
	Content        string
	Model          string
}

type Started struct {
	MessageID uuid.UUID
	Truncated bool
	Deltas    <-chan provider.Delta
}

type Service struct {
	conversations Conversations
	completer     Completer
	catalogue     Catalogue
	budget        Budget
	recaller      Recaller
	observer      Observer
	metrics       Metrics
	live          *liveRegistry
	log           *zap.Logger
}

func NewService(
	conversations Conversations,
	completer Completer,
	catalogue Catalogue,
	budget Budget,
	recaller Recaller,
	observer Observer,
	metrics Metrics,
	log *zap.Logger,
) *Service {
	return &Service{
		conversations: conversations,
		completer:     completer,
		catalogue:     catalogue,
		budget:        budget,
		recaller:      recaller,
		observer:      observer,
		metrics:       metrics,
		live:          newLiveRegistry(),
		log:           log,
	}
}

func (s *Service) Send(ctx context.Context, cmd SendCommand) (Started, error) {
	conversation, err := s.conversations.Get(ctx, cmd.ConversationID, cmd.UserID)
	if err != nil {
		return Started{}, err
	}

	modelID := cmd.Model
	if modelID == "" {
		modelID = conversation.DefaultModel
	}

	thread, err := s.conversations.Thread(ctx, conversation.ID, threadLimit)
	if err != nil {
		return Started{}, err
	}

	userMessage := model.Message{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		ParentID:       conversation.HeadMessageID,
		Role:           model.RoleUser,
		Content:        cmd.Content,
		Status:         model.MessageStatusComplete,
	}

	prompt, ok := domain.BuildPrompt(thread, s.memory(ctx, cmd.UserID), userMessage).
		Fit(s.catalogue.ContextWindow(modelID))
	if !ok {
		return Started{}, constants.ErrContextTooLarge
	}

	reservation, err := s.budget.Reserve(ctx, cmd.UserID, modelID, prompt.TokenEstimate(), defaultMaxTokens)
	if err != nil {
		s.metrics.IncQuotaRejection(rejectionReason(err))
		return Started{}, err
	}

	assistant := model.Message{
		ID:             uuid.New(),
		ConversationID: conversation.ID,
		ParentID:       &userMessage.ID,
		Role:           model.RoleAssistant,
		Status:         model.MessageStatusStreaming,
		Model:          modelID,
	}
	if err := s.conversations.AppendPair(ctx, userMessage, assistant); err != nil {
		_ = s.budget.Release(ctx, reservation)
		return Started{}, err
	}

	// The generation context is deliberately detached from the request. A user
	// losing signal mid-answer must not destroy a generation already paid for;
	// only an explicit stop cancels it.
	genCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), generationTimeout)

	upstream, err := s.completer.Stream(genCtx, toProviderRequest(prompt, modelID))
	if err != nil {
		err = fmt.Errorf("%w: %s", constants.ErrModelUnavailable, err)
		cancel()
		_ = s.budget.Release(context.WithoutCancel(ctx), reservation)
		assistant.MarkFailed(err.Error())
		_ = s.conversations.Finalize(context.WithoutCancel(ctx), assistant)
		return Started{}, err
	}

	stream := newLiveStream(cancel)
	s.live.add(assistant.ID, stream)
	subscription := stream.Subscribe()

	go s.consume(genCtx, upstream, stream, assistant, reservation, cmd.UserID, cmd.Content)

	return Started{MessageID: assistant.ID, Truncated: prompt.Truncated(), Deltas: subscription}, nil
}

// Reattach returns the live output of a generation already in flight, replaying
// what it has emitted so far.
func (s *Service) Reattach(messageID uuid.UUID) (<-chan provider.Delta, error) {
	stream, ok := s.live.get(messageID)
	if !ok {
		return nil, constants.ErrNotFound
	}
	return stream.Subscribe(), nil
}

func (s *Service) Stop(messageID uuid.UUID) error {
	stream, ok := s.live.get(messageID)
	if !ok {
		return constants.ErrNotFound
	}
	stream.Stop()
	return nil
}

// consume owns the generation for its whole lifetime, independent of whether
// any client is still listening.
func (s *Service) consume(
	ctx context.Context,
	upstream <-chan provider.Delta,
	stream *LiveStream,
	assistant model.Message,
	reservation quota.Reservation,
	userID uuid.UUID,
	userText string,
) {
	defer s.live.remove(assistant.ID)
	defer stream.close()

	// Persistence must outlive cancellation: a stopped generation still has to
	// record what it produced and settle what it spent.
	persistCtx := context.WithoutCancel(ctx)

	started := time.Now()
	firstToken := false

	var content []byte
	dirty := false

	flush := time.NewTicker(partialFlushInterval)
	defer flush.Stop()

	for {
		select {
		case <-flush.C:
			if dirty {
				if err := s.conversations.AppendPartial(persistCtx, assistant.ID, string(content)); err != nil {
					s.log.Warn("partial flush failed", zap.Error(err))
				}
				dirty = false
			}

		case delta, ok := <-upstream:
			if !ok {
				s.metrics.ObserveStream(assistant.Model, "stopped", time.Since(started))
				s.finish(persistCtx, stream, assistant, reservation, string(content), nil)
				s.observe(userID, assistant.ID, userText, string(content))
				return
			}
			stream.publish(delta)

			switch delta.Type {
			case provider.DeltaText:
				if !firstToken {
					s.metrics.ObserveFirstToken(assistant.Model, time.Since(started))
					firstToken = true
				}
				content = append(content, delta.Text...)
				dirty = true

			case provider.DeltaDone:
				s.metrics.ObserveStream(assistant.Model, "complete", time.Since(started))
				s.finish(persistCtx, stream, assistant, reservation, string(content), delta.Usage)
				s.observe(userID, assistant.ID, userText, string(content))
				return

			case provider.DeltaError:
				s.metrics.IncUpstreamError(assistant.Model)
				s.metrics.ObserveStream(assistant.Model, "failed", time.Since(started))
				assistant.Content = string(content)
				assistant.MarkFailed(delta.Err.Error())
				if err := s.conversations.Finalize(persistCtx, assistant); err != nil {
					s.log.Error("finalize failed", zap.Error(err))
				}
				if _, err := s.budget.Settle(persistCtx, reservation, assistant.ID, provider.Usage{}); err != nil {
					s.log.Error("settle failed", zap.Error(err))
				}
				return
			}
		}
	}
}

func (s *Service) finish(
	ctx context.Context,
	_ *LiveStream,
	assistant model.Message,
	reservation quota.Reservation,
	content string,
	usage *provider.Usage,
) {
	reported := provider.Usage{}
	if usage != nil {
		reported = *usage
	}

	settled, err := s.budget.Settle(ctx, reservation, assistant.ID, reported)
	if err != nil {
		s.log.Error("settle failed", zap.Error(err))
	}

	s.metrics.ObserveUsage(assistant.Model, settled.In, settled.Out, settled.CostUSDMicros)

	assistant.Content = content
	if err := assistant.MarkComplete(settled); err != nil {
		s.log.Error("mark complete failed", zap.Error(err))
	}
	if err := s.conversations.Finalize(ctx, assistant); err != nil {
		s.log.Error("finalize failed", zap.Error(err))
	}
}

func (s *Service) observe(userID, messageID uuid.UUID, userText, assistantText string) {
	if s.observer == nil || assistantText == "" {
		return
	}
	s.observer.Observe(userID, messageID, userText, assistantText)
}

func (s *Service) memory(ctx context.Context, userID uuid.UUID) domain.MemoryContext {
	if s.recaller == nil {
		return domain.MemoryContext{}
	}

	memory, err := s.recaller.Context(ctx, userID)
	if err != nil {
		s.log.Warn("memory unavailable", zap.Error(err))
		return domain.MemoryContext{}
	}
	return memory
}

func toProviderRequest(prompt domain.PromptContext, modelID string) provider.ChatRequest {
	messages := make([]provider.Message, 0, len(prompt.Messages))
	for _, m := range prompt.Messages {
		role := provider.RoleUser
		if m.Role == model.RoleAssistant {
			role = provider.RoleAssistant
		}
		messages = append(messages, provider.Message{Role: role, Content: m.Content})
	}

	return provider.ChatRequest{
		Model:     modelID,
		System:    prompt.System,
		Messages:  messages,
		MaxTokens: defaultMaxTokens,
	}
}

func rejectionReason(err error) string {
	switch {
	case errors.Is(err, constants.ErrQuotaExceeded):
		return "quota_exceeded"
	case errors.Is(err, constants.ErrModelUnavailable):
		return "model_not_allowed"
	}
	return "other"
}
