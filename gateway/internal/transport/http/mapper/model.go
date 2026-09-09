package mapper

import (
	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
)

func ToModelResponses(models []provider.ModelInfo) []dto.ModelResponse {
	out := make([]dto.ModelResponse, 0, len(models))
	for _, m := range models {
		out = append(out, dto.ModelResponse{
			ID:            m.ID,
			Provider:      m.Provider,
			Tier:          string(m.Tier),
			ContextWindow: m.ContextWindow,
			Available:     m.Available,
		})
	}
	return out
}
