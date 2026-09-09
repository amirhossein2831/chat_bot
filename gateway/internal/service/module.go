package service

import (
	"go.uber.org/fx"

	"github.com/amirhossein2831/chat_bot/gateway/internal/service/auth"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

var Module = fx.Module("service",
	fx.Provide(
		auth.NewService,
		fx.Annotate(func(s *auth.Service) *auth.Service { return s }, fx.As(new(mw.Authenticator))),
	),
)
