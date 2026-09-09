package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
)

var Module = fx.Module("transport",
	fx.Provide(
		NewRouter,
		func(m *chi.Mux) http.Handler { return m },
		NewServer,
	),
)
