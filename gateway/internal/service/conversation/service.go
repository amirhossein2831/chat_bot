package conversation

import (
	"context"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	"github.com/amirhossein2831/chat_bot/gateway/internal/domain"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

const (
	defaultListLimit   = 50
	defaultThreadLimit = 200
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID, title, defaultModel string) (model.Conversation, error) {
	conversation := model.Conversation{
		ID:           uuid.New(),
		UserID:       userID,
		Title:        title,
		DefaultModel: defaultModel,
	}
	return s.repo.Create(ctx, conversation)
}

func (s *Service) Get(ctx context.Context, id, userID uuid.UUID) (model.Conversation, error) {
	return s.repo.Get(ctx, id, userID)
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, query string) ([]model.Conversation, error) {
	if query != "" {
		return s.repo.Search(ctx, userID, query, defaultListLimit)
	}
	return s.repo.List(ctx, userID, defaultListLimit)
}

func (s *Service) Rename(ctx context.Context, id, userID uuid.UUID, title string) error {
	return s.repo.Rename(ctx, id, userID, title)
}

func (s *Service) Archive(ctx context.Context, id, userID uuid.UUID) error {
	return s.repo.Archive(ctx, id, userID)
}

// Thread returns the active branch. Ownership is verified first so an
// unrelated conversation id cannot leak its messages.
func (s *Service) Thread(ctx context.Context, id, userID uuid.UUID) (domain.Thread, error) {
	if _, err := s.repo.Get(ctx, id, userID); err != nil {
		return domain.Thread{}, err
	}

	thread, err := s.repo.Thread(ctx, id, defaultThreadLimit)
	if err != nil {
		return domain.Thread{}, err
	}

	parents := make([]uuid.UUID, 0, len(thread.Messages))
	root := false
	for _, m := range thread.Messages {
		if m.ParentID == nil {
			root = true
			continue
		}
		parents = append(parents, *m.ParentID)
	}

	siblings, err := s.repo.SiblingSets(ctx, id, parents, root)
	if err != nil {
		return domain.Thread{}, err
	}
	return thread.WithSiblings(siblings), nil
}

// SwitchBranch moves the conversation head to a sibling of the current one,
// which is how a regenerated answer is chosen without deleting the other.
func (s *Service) SwitchBranch(ctx context.Context, conversationID, userID, messageID uuid.UUID) error {
	conversation, err := s.repo.Get(ctx, conversationID, userID)
	if err != nil {
		return err
	}

	thread, err := s.repo.Thread(ctx, conversation.ID, defaultThreadLimit)
	if err != nil {
		return err
	}
	if !containsMessage(thread, messageID) {
		if err := s.verifySibling(ctx, conversation.ID, thread, messageID); err != nil {
			return err
		}
	}

	head, err := s.repo.DeepestDescendant(ctx, messageID)
	if err != nil {
		return err
	}
	return s.repo.SetHead(ctx, conversationID, head)
}

// verifySibling accepts any branch that forks off the path being viewed, not
// only one hanging beside its last turn: editing a question halfway up the
// thread forks there, and that branch has to remain reachable. A fork at the
// very start has no parent at all, which is a sibling relationship too.
func (s *Service) verifySibling(
	ctx context.Context,
	conversationID uuid.UUID,
	thread domain.Thread,
	messageID uuid.UUID,
) error {
	target, err := s.repo.Message(ctx, messageID)
	if err != nil {
		return err
	}
	if target.ConversationID != conversationID {
		return constants.ErrNotFound
	}

	for _, m := range thread.Messages {
		if samePosition(m.ParentID, target.ParentID) {
			return nil
		}
	}
	return constants.ErrNotFound
}

func samePosition(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func containsMessage(thread domain.Thread, id uuid.UUID) bool {
	for _, m := range thread.Messages {
		if m.ID == id {
			return true
		}
	}
	return false
}
