package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	httptransport "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/controller/auth"
	chatctrl "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/controller/chat"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/controller/conversation"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/controller/health"
	modelctrl "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/controller/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"
)

var Module = fx.Module("transport",
	fx.Provide(
		httptransport.NewServer,
		router.New,
		func(mux *chi.Mux) http.Handler { return mux },

		fx.Annotate(auth.NewController,
			fx.As(new(router.Registrar)),
			fx.ResultTags(`group:"routes"`),
		),
		fx.Annotate(chatctrl.NewController,
			fx.As(new(router.Registrar)),
			fx.ResultTags(`group:"routes"`),
		),
		fx.Annotate(conversation.NewController,
			fx.As(new(router.Registrar)),
			fx.ResultTags(`group:"routes"`),
		),
		fx.Annotate(modelctrl.NewController,
			fx.As(new(router.Registrar)),
			fx.ResultTags(`group:"routes"`),
		),
		fx.Annotate(health.NewController,
			fx.As(new(router.Registrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),
)
