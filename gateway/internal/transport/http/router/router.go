package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

const apiBasePath = "/api/v1"

type Params struct {
	fx.In

	Log        *zap.Logger
	Auth       mw.Authenticator
	Metrics    mw.Observer
	Gatherer   prometheus.Gatherer
	Registrars []Registrar `group:"routes"`
}

func New(p Params) *chi.Mux {
	root := chi.NewRouter()
	root.Use(mw.RequestID, mw.Recover(p.Log, httpio.Error), mw.Logging(p.Log, p.Metrics))

	root.Handle("/metrics", promhttp.HandlerFor(p.Gatherer, promhttp.HandlerOpts{}))

	root.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		httpio.Error(w, constants.ErrNotFound)
	})
	root.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		httpio.Error(w, constants.ErrMethodNotAllowed)
	})

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
