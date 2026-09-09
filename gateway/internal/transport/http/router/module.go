package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

var Module = fx.Module("router",
	fx.Provide(
		New,
		func(mux *chi.Mux) http.Handler { return mux },
	),
)
