package controller

import (
	"go.uber.org/fx"

	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/controller/auth"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"
)

var Module = fx.Module("controller",
	fx.Provide(
		fx.Annotate(auth.NewController,
			fx.As(new(router.Registrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),
)
