package catalogue

import (
	"context"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

type Service struct {
	registry Registry
}

func NewService(registry Registry) *Service {
	return &Service{registry: registry}
}

// Available is where plan-based filtering will live: a free user should not be
// offered a frontier model at all, rather than being refused after choosing one.
func (s *Service) Available(_ context.Context, _ model.User) []provider.ModelInfo {
	return s.registry.Models()
}

func (s *Service) ContextWindow(modelID string) int {
	return s.registry.ContextWindow(modelID)
}
