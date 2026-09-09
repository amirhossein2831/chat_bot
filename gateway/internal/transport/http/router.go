package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"

	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

const apiBasePath = "/api/v1"

type RouterParams struct {
	fx.In

	Log        *zap.Logger
	Auth       mw.Authenticator
	Registrars []Registrar `group:"routes"`
}

func NewRouter(p RouterParams) *chi.Mux {
	r := chi.NewRouter()
	r.Use(mw.RequestID, mw.Recover(p.Log))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route(apiBasePath, func(api chi.Router) {
		public := api.Group(nil)

		protected := api.Group(nil)
		protected.Use(mw.RequireAuth(p.Auth, WriteError))

		for _, registrar := range p.Registrars {
			registrar.Routes(public, protected)
		}
	})

	return r
}
