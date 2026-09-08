package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

// Catches missing, duplicated or ambiguous fx providers in CI rather than as a
// panic on production startup, which is fx's main cost without this test.
func TestApplicationGraphIsValid(t *testing.T) {
	t.Setenv("APP_POSTGRES_DSN", "postgres://u:p@localhost:5432/db")
	t.Setenv("APP_REDIS_ADDR", "localhost:6379")

	require.NoError(t, fx.ValidateApp(options()))
}
