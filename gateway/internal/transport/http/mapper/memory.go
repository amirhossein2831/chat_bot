package mapper

import (
	"time"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
)

func ToFactResponse(fact model.Fact) dto.FactResponse {
	return dto.FactResponse{
		ID:        fact.ID.String(),
		Text:      fact.Text,
		Pinned:    fact.Pinned,
		CreatedAt: fact.CreatedAt.Format(time.RFC3339),
	}
}

func ToFactResponses(facts []model.Fact) []dto.FactResponse {
	out := make([]dto.FactResponse, 0, len(facts))
	for _, fact := range facts {
		out = append(out, ToFactResponse(fact))
	}
	return out
}
