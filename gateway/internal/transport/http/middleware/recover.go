package middleware

import (
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

func Recover(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				v := recover()
				if v == nil {
					return
				}
				log.Error("panic recovered",
					zap.Any("panic", v),
					zap.String("request_id", RequestIDFrom(r.Context())),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
					zap.ByteString("stack", debug.Stack()))
				w.WriteHeader(http.StatusInternalServerError)
			}()
			next.ServeHTTP(w, r)
		})
	}
}
