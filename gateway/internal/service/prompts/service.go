package prompts

import (
	"context"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

const defaultLocale = "fa"

type Repository interface {
	List(ctx context.Context, locale, category string) ([]model.Prompt, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, locale, category string) ([]model.Prompt, error) {
	if locale != "fa" && locale != "en" {
		locale = defaultLocale
	}
	return s.repo.List(ctx, locale, category)
}
