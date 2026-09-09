package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"

	httptransport "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/controller/auth"
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
	),
)
