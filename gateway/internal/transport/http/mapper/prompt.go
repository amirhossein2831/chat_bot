package mapper

import (
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
)

func ToPromptResponses(prompts []model.Prompt) []dto.PromptResponse {
	out := make([]dto.PromptResponse, 0, len(prompts))
	for _, p := range prompts {
		out = append(out, dto.PromptResponse{
			ID:       p.ID.String(),
			Locale:   p.Locale,
			Category: p.Category,
			Title:    p.Title,
			Body:     p.Body,
		})
	}
	return out
}
