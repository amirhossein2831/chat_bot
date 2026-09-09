package router

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

const apiBasePath = "/api/v1"

type Params struct {
	fx.In

	Log        *zap.Logger
	Auth       mw.Authenticator
	Registrars []Registrar `group:"routes"`
}

func New(p Params) *chi.Mux {
	root := chi.NewRouter()
	root.Use(mw.RequestID, mw.Recover(p.Log))

	root.Route(apiBasePath, func(api chi.Router) {
		public := api.Group(nil)

		protected := api.Group(nil)
		protected.Use(mw.RequireAuth(p.Auth, httpio.Error))

		scopes := Routers{Root: root, Public: public, Protected: protected}
		for _, registrar := range p.Registrars {
			registrar.Routes(scopes)
		}
	})

	return root
}
