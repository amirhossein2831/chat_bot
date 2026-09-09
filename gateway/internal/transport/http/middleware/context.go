package middleware

import "context"

type ctxKey int

const (
	requestIDKey ctxKey = iota
	userKey
)

func valueFrom[T any](ctx context.Context, key ctxKey) T {
	v, _ := ctx.Value(key).(T)
	return v
}

func RequestIDFrom(ctx context.Context) string {
	return valueFrom[string](ctx, requestIDKey)
}
