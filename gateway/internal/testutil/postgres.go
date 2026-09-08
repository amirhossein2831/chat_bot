// Package testutil provides integration-test fixtures. It is only imported
// from _test.go files.
package testutil

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // database/sql driver for goose
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// defaultImage is pinned to a tag available locally. Docker Hub is not
// reachable from here and the configured pull-through mirror refuses, so the
// image must already be present. Override with TEST_POSTGRES_IMAGE where a
// registry is available.
const defaultImage = "postgres:16-alpine"

// NewPostgres starts a disposable Postgres, applies every migration, and
// returns a pool. The container is removed when the test finishes.
func NewPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	// Ryuk is testcontainers' reaper container, and its image cannot be pulled
	// here. Containers are terminated explicitly in the cleanup below, so the
	// reaper is not needed.
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	image := os.Getenv("TEST_POSTGRES_IMAGE")
	if image == "" {
		image = defaultImage
	}

	container, err := tcpostgres.Run(ctx, image,
		tcpostgres.WithDatabase("app"),
		tcpostgres.WithUsername("app"),
		tcpostgres.WithPassword("app"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(90*time.Second)),
	)
	require.NoError(t, err, "starting postgres container")
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = container.Terminate(ctx)
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	applyMigrations(t, dsn)

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	return pool
}

func applyMigrations(t *testing.T, dsn string) {
	t.Helper()

	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	goose.SetLogger(goose.NopLogger())
	require.NoError(t, goose.SetDialect("postgres"))
	require.NoError(t, goose.Up(db, migrationsDir(t)))
}

// migrationsDir walks up from the working directory to find migrations/.
// A fixed relative path would break the moment this helper is called from a
// package at a different depth.
func migrationsDir(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		candidate := filepath.Join(dir, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("no migrations directory found above %s", dir)
		}
		dir = parent
	}
}
