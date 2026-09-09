package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeConfigs(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600))
	}
	return dir
}

const baseYAML = `
http:
  addr: ":8080"
postgres:
  dsn: "postgres://base@localhost:5432/db"
  max_conns: 10
redis:
  addr: "localhost:6379"
`

func TestDefaultsToLocalEnvironment(t *testing.T) {
	dir := writeConfigs(t, map[string]string{"config.yaml": baseYAML})

	cfg, err := LoadFrom(dir)

	require.NoError(t, err)
	require.Equal(t, "local", cfg.Env)
}

func TestEnvironmentFileOverridesBase(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	dir := writeConfigs(t, map[string]string{
		"config.yaml": baseYAML,
		"config.production.yaml": `
http:
  addr: ":80"
postgres:
  max_conns: 50
`,
	})

	cfg, err := LoadFrom(dir)

	require.NoError(t, err)
	require.Equal(t, "production", cfg.Env)
	require.Equal(t, ":80", cfg.HTTP.Addr, "environment file must override base")
	require.Equal(t, int32(50), cfg.Postgres.MaxConns)
	require.Equal(t, "postgres://base@localhost:5432/db", cfg.Postgres.DSN,
		"values absent from the environment file must fall through to base")
}

func TestEnvVarBeatsEnvironmentFile(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_HTTP_ADDR", ":9999")
	dir := writeConfigs(t, map[string]string{
		"config.yaml":            baseYAML,
		"config.production.yaml": "http:\n  addr: \":80\"\n",
	})

	cfg, err := LoadFrom(dir)

	require.NoError(t, err)
	require.Equal(t, ":9999", cfg.HTTP.Addr)
}

func TestSecretsCanComeFromEnvironmentAlone(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_POSTGRES_DSN", "postgres://secret@prod:5432/db")
	t.Setenv("APP_REDIS_ADDR", "redis-prod:6379")
	dir := writeConfigs(t, map[string]string{
		"config.yaml": "http:\n  addr: \":8080\"\n",
	})

	cfg, err := LoadFrom(dir)

	require.NoError(t, err)
	require.Equal(t, "postgres://secret@prod:5432/db", cfg.Postgres.DSN)
}

func TestMissingFilesAreToleratedWhenEnvironmentIsComplete(t *testing.T) {
	t.Setenv("APP_POSTGRES_DSN", "postgres://u:p@localhost:5432/db")
	t.Setenv("APP_REDIS_ADDR", "localhost:6379")

	cfg, err := LoadFrom(t.TempDir())

	require.NoError(t, err)
	require.Equal(t, ":8080", cfg.HTTP.Addr)
}

func TestRejectsUnknownEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "wharrgarbl")
	dir := writeConfigs(t, map[string]string{"config.yaml": baseYAML})

	_, err := LoadFrom(dir)

	require.Error(t, err)
}

func TestRejectsMissingRequiredValues(t *testing.T) {
	dir := writeConfigs(t, map[string]string{"config.yaml": "http:\n  addr: \":8080\"\n"})

	_, err := LoadFrom(dir)

	require.Error(t, err, "a missing postgres DSN must fail at boot")
}

func TestAppliesDefaults(t *testing.T) {
	dir := writeConfigs(t, map[string]string{"config.yaml": baseYAML})

	cfg, err := LoadFrom(dir)

	require.NoError(t, err)
	require.Equal(t, 30, cfg.HTTP.ShutdownTimeout)
	require.Equal(t, 0, cfg.Redis.DB)
}

func TestEnvVarNamingConvention(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("APP_HTTP_ADDR", ":7000")
	t.Setenv("APP_HTTP_SHUTDOWN_TIMEOUT", "45")
	t.Setenv("APP_POSTGRES_DSN", "postgres://env@host:5432/db")
	t.Setenv("APP_POSTGRES_MAX_CONNS", "77")
	t.Setenv("APP_REDIS_ADDR", "redis:6379")
	t.Setenv("APP_REDIS_DB", "3")

	cfg, err := LoadFrom(t.TempDir())

	require.NoError(t, err)
	require.Equal(t, "staging", cfg.Env)
	require.Equal(t, ":7000", cfg.HTTP.Addr)
	require.Equal(t, 45, cfg.HTTP.ShutdownTimeout)
	require.Equal(t, "postgres://env@host:5432/db", cfg.Postgres.DSN)
	require.Equal(t, int32(77), cfg.Postgres.MaxConns)
	require.Equal(t, "redis:6379", cfg.Redis.Addr)
	require.Equal(t, 3, cfg.Redis.DB)
}
