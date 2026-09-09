package controller

import (
	"go.uber.org/fx"

	httptransport "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/controller/auth"
)

var Module = fx.Module("controller",
	fx.Provide(
		fx.Annotate(auth.NewController,
			fx.As(new(httptransport.Registrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),
)
