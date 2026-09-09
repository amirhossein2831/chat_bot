package middleware

import (
	"context"
	"net/http"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

const sessionCookie = "session"

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (model.User, error)
}

func RequireAuth(auth Authenticator, writeErr func(http.ResponseWriter, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookie)
			if err != nil {
				writeErr(w, constants.ErrUnauthorized)
				return
			}

			user, err := auth.Authenticate(r.Context(), cookie.Value)
			if err != nil {
				writeErr(w, err)
				return
			}

			if s := scopeFrom(r.Context()); s != nil {
				s.UserID = user.ID.String()
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, user)))
		})
	}
}

func UserFrom(ctx context.Context) model.User {
	return valueFrom[model.User](ctx, userKey)
}
