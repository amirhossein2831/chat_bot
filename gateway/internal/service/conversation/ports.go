package conversation

import (
	"context"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/domain"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

type Repository interface {
	Create(ctx context.Context, c model.Conversation) (model.Conversation, error)
	Get(ctx context.Context, id, userID uuid.UUID) (model.Conversation, error)
	List(ctx context.Context, userID uuid.UUID, limit int) ([]model.Conversation, error)
	Search(ctx context.Context, userID uuid.UUID, query string, limit int) ([]model.Conversation, error)
	Rename(ctx context.Context, id, userID uuid.UUID, title string) error
	Archive(ctx context.Context, id, userID uuid.UUID) error
	Thread(ctx context.Context, conversationID uuid.UUID, limit int) (domain.Thread, error)
	Siblings(ctx context.Context, parentID uuid.UUID) ([]model.Message, error)
	SetHead(ctx context.Context, conversationID, messageID uuid.UUID) error
}
