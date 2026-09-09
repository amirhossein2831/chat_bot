package middleware

import "context"

type ctxKey int

const (
	requestIDKey ctxKey = iota
	userKey
	scopeKey
)

// scope is filled in by inner middleware and read by outer middleware after
// the handler returns. Context values set on a derived request are invisible
// to an outer wrapper, so the shared pointer is what carries them back out.
type scope struct {
	UserID string
}

func scopeFrom(ctx context.Context) *scope {
	s, _ := ctx.Value(scopeKey).(*scope)
	return s
}

func valueFrom[T any](ctx context.Context, key ctxKey) T {
	v, _ := ctx.Value(key).(T)
	return v
}

func RequestIDFrom(ctx context.Context) string {
	return valueFrom[string](ctx, requestIDKey)
}
