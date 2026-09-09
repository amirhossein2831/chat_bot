package controller

import (
	"go.uber.org/fx"

	httptransport "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http"
)

var Module = fx.Module("controller",
	fx.Provide(
		fx.Annotate(NewAuth,
			fx.As(new(httptransport.Registrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),
)
