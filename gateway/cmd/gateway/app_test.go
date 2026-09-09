package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestApplicationGraphIsValid(t *testing.T) {
	t.Setenv("APP_POSTGRES_DSN", "postgres://u:p@localhost:5432/db")
	t.Setenv("APP_REDIS_ADDR", "localhost:6379")

	require.NoError(t, fx.ValidateApp(options()))
}
