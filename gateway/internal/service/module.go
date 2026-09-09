package service

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/registry"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/auth"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/catalogue"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/chat"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/conversation"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/quota"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

var Module = fx.Module("service",
	fx.Provide(
		auth.NewService,
		conversation.NewService,
		quota.NewService,
		func(
			conversations chat.Conversations,
			registry *registry.Registry,
			catalogueService *catalogue.Service,
			budget *quota.Service,
			log *zap.Logger,
		) *chat.Service {
			return chat.NewService(conversations, registry, catalogueService, budget, log)
		},
		func(r *registry.Registry) *catalogue.Service { return catalogue.NewService(r) },
		fx.Annotate(func(s *auth.Service) *auth.Service { return s }, fx.As(new(mw.Authenticator))),
	),
)
