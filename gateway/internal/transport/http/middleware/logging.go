package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Observer interface {
	ObserveRequest(method, route, status string, d time.Duration)
}

type recorder struct {
	http.ResponseWriter
	status int
}

func (r *recorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Unwrap lets http.ResponseController reach the real writer, without which
// streaming responses lose their flusher.
func (r *recorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// Logging records one line per request. Message content is never included:
// request, user and conversation ids are the only identifiers logged.
func Logging(log *zap.Logger, metrics Observer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			rec := &recorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r)

			elapsed := time.Since(started)
			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = "unmatched"
			}
			status := strconv.Itoa(rec.status)

			if metrics != nil {
				metrics.ObserveRequest(r.Method, route, status, elapsed)
			}

			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("route", route),
				zap.Int("status", rec.status),
				zap.Duration("took", elapsed),
				zap.String("request_id", RequestIDFrom(r.Context())),
			}
			if s := scopeFrom(r.Context()); s != nil && s.UserID != "" {
				fields = append(fields, zap.String("user_id", s.UserID))
			}

			if rec.status >= http.StatusInternalServerError {
				log.Error("request failed", fields...)
				return
			}
			log.Info("request", fields...)
		})
	}
}
