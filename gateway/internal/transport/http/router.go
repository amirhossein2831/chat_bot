package http

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

func NewRouter(log *slog.Logger) *chi.Mux {
	r := chi.NewRouter()
	r.Use(mw.RequestID, mw.Recover(log))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return r
}
