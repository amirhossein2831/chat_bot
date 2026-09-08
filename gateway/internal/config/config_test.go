package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadUsesEnvOverrides(t *testing.T) {
	t.Setenv("APP_HTTP_ADDR", ":9999")
	t.Setenv("APP_POSTGRES_DSN", "postgres://u:p@localhost:5432/db")
	t.Setenv("APP_REDIS_ADDR", "localhost:6379")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, ":9999", cfg.HTTP.Addr)
}

// Every key must have a default registered, or viper never consults the
// environment for it during Unmarshal and the override is silently ignored.
func TestLoadReadsEveryRequiredKeyFromEnv(t *testing.T) {
	t.Setenv("APP_HTTP_ADDR", ":8080")
	t.Setenv("APP_POSTGRES_DSN", "postgres://u:p@localhost:5432/db")
	t.Setenv("APP_REDIS_ADDR", "redis-host:6379")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "postgres://u:p@localhost:5432/db", cfg.Postgres.DSN)
	require.Equal(t, "redis-host:6379", cfg.Redis.Addr)
}

func TestLoadRejectsMissingRequiredValues(t *testing.T) {
	t.Setenv("APP_HTTP_ADDR", ":8080")
	t.Setenv("APP_POSTGRES_DSN", "")
	t.Setenv("APP_REDIS_ADDR", "localhost:6379")

	_, err := Load()
	require.Error(t, err)
}

func TestLoadAppliesDefaults(t *testing.T) {
	t.Setenv("APP_POSTGRES_DSN", "postgres://u:p@localhost:5432/db")
	t.Setenv("APP_REDIS_ADDR", "localhost:6379")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "dev", cfg.Env)
	require.Equal(t, ":8080", cfg.HTTP.Addr)
	require.Equal(t, 30, cfg.HTTP.ShutdownTimeout)
	require.Equal(t, int32(10), cfg.Postgres.MaxConns)
}
