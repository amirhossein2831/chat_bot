package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const requestIDHeader = "X-Request-ID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set(requestIDHeader, id)

		ctx := context.WithValue(r.Context(), requestIDKey, id)
		ctx = context.WithValue(ctx, scopeKey, &scope{})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
