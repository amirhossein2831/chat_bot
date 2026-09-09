package mapper

import (
	"time"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/domain"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
)

func ToConversationResponse(c model.Conversation) dto.ConversationResponse {
	return dto.ConversationResponse{
		ID:            c.ID.String(),
		Title:         c.Title,
		DefaultModel:  c.DefaultModel,
		HeadMessageID: uuidPtrToString(c.HeadMessageID),
		Archived:      c.IsArchived(),
		CreatedAt:     c.CreatedAt.Format(time.RFC3339),
	}
}

func ToConversationResponses(conversations []model.Conversation) []dto.ConversationResponse {
	out := make([]dto.ConversationResponse, 0, len(conversations))
	for _, c := range conversations {
		out = append(out, ToConversationResponse(c))
	}
	return out
}

func ToMessageResponse(m model.Message) dto.MessageResponse {
	return dto.MessageResponse{
		ID:        m.ID.String(),
		ParentID:  uuidPtrToString(m.ParentID),
		Role:      string(m.Role),
		Content:   m.Content,
		Status:    string(m.Status),
		Model:     m.Model,
		TokensIn:  m.TokensIn,
		TokensOut: m.TokensOut,
		Error:     m.Error,
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}
}

func ToThreadResponse(t domain.Thread) dto.ThreadResponse {
	messages := make([]dto.MessageResponse, 0, t.Len())
	for _, m := range t.Messages {
		response := ToMessageResponse(m)
		for _, sibling := range t.Siblings[m.ID] {
			response.Siblings = append(response.Siblings, sibling.String())
		}
		messages = append(messages, response)
	}
	return dto.ThreadResponse{Messages: messages}
}

func uuidPtrToString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	value := id.String()
	return &value
}
