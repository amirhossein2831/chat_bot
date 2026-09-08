package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/amirhossein2831/chat_bot/gateway/internal/testutil"
)

func TestHarnessAppliesMigrations(t *testing.T) {
	pool := testutil.NewPostgres(t)

	var n int
	err := pool.QueryRow(context.Background(),
		`select count(*) from information_schema.tables
		  where table_schema = 'public' and table_name in ('users', 'invites')`,
	).Scan(&n)

	require.NoError(t, err)
	require.Equal(t, 2, n)
}

// The extensions are what make citext emails and trigram Persian search work.
// A migration that silently failed to create them would only surface much
// later, as case-sensitive logins or a sequential scan on every search.
func TestHarnessEnablesRequiredExtensions(t *testing.T) {
	pool := testutil.NewPostgres(t)

	rows, err := pool.Query(context.Background(),
		`select extname from pg_extension where extname in ('citext', 'pg_trgm')`)
	require.NoError(t, err)
	defer rows.Close()

	var found []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		found = append(found, name)
	}
	require.ElementsMatch(t, []string{"citext", "pg_trgm"}, found)
}

func TestEachTestGetsAnIsolatedDatabase(t *testing.T) {
	ctx := context.Background()

	first := testutil.NewPostgres(t)
	_, err := first.Exec(ctx, `insert into invites (code) values ('LEAKED')`)
	require.NoError(t, err)

	second := testutil.NewPostgres(t)
	var n int
	require.NoError(t, second.QueryRow(ctx,
		`select count(*) from invites where code = 'LEAKED'`).Scan(&n))
	require.Zero(t, n, "state must not leak between harness instances")
}
