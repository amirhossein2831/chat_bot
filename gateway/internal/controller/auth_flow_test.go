package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/controller"
	pgrepo "github.com/amirhossein2831/chat_bot/gateway/internal/repository/postgres"
	redisrepo "github.com/amirhossein2831/chat_bot/gateway/internal/repository/redis"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/auth"
	httptransport "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http"
	"github.com/amirhossein2831/chat_bot/gateway/internal/util/testutil"
)

type client struct {
	t    *testing.T
	base string
	http *http.Client
	pool *pgxpool.Pool
}

func newStack(t *testing.T) *client {
	t.Helper()

	pool := testutil.NewPostgres(t)
	rdb := testutil.NewRedis(t)

	service := auth.NewService(
		pgrepo.NewUserRepository(pool),
		redisrepo.NewSessionRepository(rdb, time.Hour),
	)
	ctrl := controller.NewAuth(service, &config.Config{
		Env:  config.EnvLocal,
		Auth: config.AuthConfig{SessionTTLHours: 1},
	})

	router := httptransport.NewRouter(httptransport.RouterParams{
		Log:        zap.NewNop(),
		Auth:       service,
		Registrars: []httptransport.Registrar{ctrl},
	})

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	return &client{t: t, base: srv.URL, http: &http.Client{Jar: jar}, pool: pool}
}

func (c *client) seedInvite(code string) {
	c.t.Helper()
	_, err := c.pool.Exec(context.Background(),
		`insert into invites (code) values ($1)`, code)
	require.NoError(c.t, err)
}

func (c *client) do(method, path string, body any) *http.Response {
	c.t.Helper()

	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		require.NoError(c.t, err)
		reader = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, c.base+path, reader)
	require.NoError(c.t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	require.NoError(c.t, err)
	c.t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func (c *client) register(email, password, invite string) *http.Response {
	return c.do(http.MethodPost, "/api/v1/auth/register", map[string]string{
		"email": email, "password": password, "invite": invite,
	})
}

func (c *client) login(email, password string) *http.Response {
	return c.do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email": email, "password": password,
	})
}

func bodyOf(t *testing.T, resp *http.Response) string {
	t.Helper()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(raw)
}

func TestRegisterLoginThenReadOwnProfile(t *testing.T) {
	c := newStack(t)
	c.seedInvite("GOOD")

	registered := c.register("a@b.com", "correct horse battery", "GOOD")
	require.Equal(t, http.StatusCreated, registered.StatusCode, bodyOf(t, registered))

	loggedIn := c.login("a@b.com", "correct horse battery")
	require.Equal(t, http.StatusOK, loggedIn.StatusCode)

	me := c.do(http.MethodGet, "/api/v1/me", nil)
	require.Equal(t, http.StatusOK, me.StatusCode)
	require.Contains(t, bodyOf(t, me), "a@b.com")
}

func TestLogoutEndsTheSession(t *testing.T) {
	c := newStack(t)
	c.seedInvite("GOOD")
	c.register("a@b.com", "correct horse battery", "GOOD")
	c.login("a@b.com", "correct horse battery")

	out := c.do(http.MethodPost, "/api/v1/auth/logout", nil)
	require.Equal(t, http.StatusNoContent, out.StatusCode)

	me := c.do(http.MethodGet, "/api/v1/me", nil)
	require.Equal(t, http.StatusUnauthorized, me.StatusCode,
		"the session must be gone from the store, not just the cookie")
}

func TestProtectedRoutesRejectAnonymousCallers(t *testing.T) {
	c := newStack(t)

	me := c.do(http.MethodGet, "/api/v1/me", nil)
	require.Equal(t, http.StatusUnauthorized, me.StatusCode)
}

func TestAnInviteIsConsumedExactlyOnce(t *testing.T) {
	c := newStack(t)
	c.seedInvite("ONCE")

	first := c.register("a@b.com", "correct horse battery", "ONCE")
	require.Equal(t, http.StatusCreated, first.StatusCode)

	second := c.register("c@d.com", "correct horse battery", "ONCE")
	require.Equal(t, http.StatusUnauthorized, second.StatusCode)
}

func TestDuplicateEmailIsRejected(t *testing.T) {
	c := newStack(t)
	c.seedInvite("ONE")
	c.seedInvite("TWO")

	require.Equal(t, http.StatusCreated, c.register("a@b.com", "correct horse battery", "ONE").StatusCode)

	duplicate := c.register("A@B.com", "correct horse battery", "TWO")
	require.Equal(t, http.StatusConflict, duplicate.StatusCode,
		"citext makes emails case-insensitive, so this is the same address")
}

func TestWrongPasswordAndUnknownEmailAreIndistinguishable(t *testing.T) {
	c := newStack(t)
	c.seedInvite("GOOD")
	c.register("a@b.com", "correct horse battery", "GOOD")

	wrongPassword := c.login("a@b.com", "wrong password here")
	unknownEmail := c.login("nobody@b.com", "correct horse battery")

	require.Equal(t, http.StatusUnauthorized, wrongPassword.StatusCode)
	require.Equal(t, http.StatusUnauthorized, unknownEmail.StatusCode)
	require.Equal(t, bodyOf(t, wrongPassword), bodyOf(t, unknownEmail),
		"a different response would let an attacker enumerate registered addresses")
}

func TestPasswordIsNeverStoredInPlaintext(t *testing.T) {
	c := newStack(t)
	c.seedInvite("GOOD")
	c.register("a@b.com", "correct horse battery", "GOOD")

	var hash string
	require.NoError(t, c.pool.QueryRow(context.Background(),
		`select password_hash from users where email = 'a@b.com'`).Scan(&hash))

	require.NotContains(t, hash, "correct horse battery")
	require.Contains(t, hash, "argon2id")
}

func TestShortPasswordIsRejectedBeforeReachingTheDatabase(t *testing.T) {
	c := newStack(t)
	c.seedInvite("GOOD")

	resp := c.register("a@b.com", "short", "GOOD")
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var n int
	require.NoError(t, c.pool.QueryRow(context.Background(),
		`select count(*) from invites where used_by is not null`).Scan(&n))
	require.Zero(t, n, "a rejected registration must not consume the invite")
}
