package service

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure/objectstore"
	"github.com/amirhossein2831/chat_bot/gateway/internal/jobs"
	"github.com/amirhossein2831/chat_bot/gateway/internal/observability"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/ocr"
	"github.com/amirhossein2831/chat_bot/gateway/internal/client/registry"
	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure/pdf"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/auth"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/catalogue"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/chat"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/conversation"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/files"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/memory"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/prompts"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/quota"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

var Module = fx.Module("service",
	fx.Provide(
		auth.NewService,
		conversation.NewService,
		quota.NewService,
		func(
			repo files.Repository,
			store *objectstore.Store,
			queue *jobs.Queue,
			pages *pdf.Reader,
			cfg *config.Config,
			log *zap.Logger,
		) *files.Service {
			return files.NewService(repo, store, queue, pages, recognizer(cfg, log), cfg, log)
		},
		prompts.NewService,
		fx.Annotate(func(s *files.Service) *files.Service { return s }, fx.As(new(jobs.Extractor))),
		func(repo memory.Repository, registry *registry.Registry, cfg *config.Config, log *zap.Logger) *memory.Service {
			return memory.NewService(repo, registry, localModelID(cfg), log)
		},
		func(
			conversations chat.Conversations,
			registry *registry.Registry,
			catalogueService *catalogue.Service,
			budget *quota.Service,
			fileService *files.Service,
			memoryService *memory.Service,
			metrics *observability.Metrics,
			log *zap.Logger,
		) *chat.Service {
			return chat.NewService(conversations, registry, catalogueService, budget,
				fileService, memoryService, memoryService, metrics, log)
		},
		func(r *registry.Registry) *catalogue.Service { return catalogue.NewService(r) },
		fx.Annotate(func(s *auth.Service) *auth.Service { return s }, fx.As(new(mw.Authenticator))),
	),
)

// localModelID picks the model used for background work such as fact
// extraction. It is always a local one so that work never spends foreign
// currency.
func localModelID(cfg *config.Config) string {
	for _, m := range cfg.Models {
		if m.Tier == "local" {
			return m.ID
		}
	}
	return ""
}

// recognizer returns nil when no OCR service is configured. A nil recogniser
// is a working state, not a broken one: scans are then reported as unsupported
// and every other format still extracts.
func recognizer(cfg *config.Config, log *zap.Logger) files.Recognizer {
	if cfg.OCR.BaseURL == "" {
		log.Info("no ocr service configured, scanned documents will be unsupported")
		return nil
	}
	return ocr.New(cfg.OCR.BaseURL, cfg.OCR.Timeout())
}
