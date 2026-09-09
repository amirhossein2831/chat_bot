package http_test

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	httptransport "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

func newTestServer(t *testing.T, logTo io.Writer) *httptest.Server {
	t.Helper()

	if logTo == nil {
		logTo = io.Discard
	}
	log := slog.New(slog.NewJSONHandler(logTo, nil))

	r := httptransport.NewRouter(log)
	r.Route("/probe", func(r chi.Router) {
		r.Get("/fail/{kind}", func(w http.ResponseWriter, req *http.Request) {
			httptransport.WriteError(w, wrapped(chi.URLParam(req, "kind")))
		})
		r.Get("/panic", func(http.ResponseWriter, *http.Request) {
			panic("boom")
		})
		r.Get("/echo-request-id", func(w http.ResponseWriter, req *http.Request) {
			httptransport.WriteJSON(w, http.StatusOK,
				map[string]string{"id": mw.RequestIDFrom(req.Context())})
		})
	})

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func wrapped(kind string) error {
	byKind := map[string]error{
		"not_found":         constants.ErrNotFound,
		"unauthorized":      constants.ErrUnauthorized,
		"forbidden":         constants.ErrForbidden,
		"conflict":          constants.ErrConflict,
		"bad_request":       constants.ErrBadRequest,
		"invalid_state":     constants.ErrInvalidTransition,
		"quota":             constants.ErrQuotaExceeded,
		"model_unavailable": constants.ErrModelUnavailable,
		"too_large":         constants.ErrContextTooLarge,
	}
	if err, ok := byKind[kind]; ok {
		return fmt.Errorf("loading thread: %w", err)
	}
	return fmt.Errorf("pq: password authentication failed for user app")
}

func get(t *testing.T, srv *httptest.Server, path string, headers map[string]string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	require.NoError(t, err)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestHealthzReportsOK(t *testing.T) {
	resp := get(t, newTestServer(t, nil), "/healthz", nil)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
}

func TestSentinelErrorsReachTheClientAsStatusCodes(t *testing.T) {
	srv := newTestServer(t, nil)

	for kind, want := range map[string]int{
		"not_found":         http.StatusNotFound,
		"unauthorized":      http.StatusUnauthorized,
		"forbidden":         http.StatusForbidden,
		"conflict":          http.StatusConflict,
		"bad_request":       http.StatusBadRequest,
		"invalid_state":     http.StatusConflict,
		"quota":             http.StatusTooManyRequests,
		"model_unavailable": http.StatusServiceUnavailable,
		"too_large":         http.StatusRequestEntityTooLarge,
	} {
		t.Run(kind, func(t *testing.T) {
			resp := get(t, srv, "/probe/fail/"+kind, nil)
			require.Equal(t, want, resp.StatusCode)
		})
	}
}

func TestUnmappedErrorLeaksNothing(t *testing.T) {
	resp := get(t, newTestServer(t, nil), "/probe/fail/unknown", nil)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	require.NotContains(t, string(body), "password")
	require.NotContains(t, string(body), "pq:")
}

func TestPanicBecomesFiveHundredAndServerKeepsServing(t *testing.T) {
	var logs syncBuffer
	srv := newTestServer(t, &logs)

	panicked := get(t, srv, "/probe/panic", nil)
	require.Equal(t, http.StatusInternalServerError, panicked.StatusCode)

	after := get(t, srv, "/healthz", nil)
	require.Equal(t, http.StatusOK, after.StatusCode, "a panic must not take the server down")

	require.Contains(t, logs.String(), "panic recovered")
}

func TestRequestIDIsGeneratedEchoedAndReachesHandlers(t *testing.T) {
	srv := newTestServer(t, nil)

	generated := get(t, srv, "/echo-request-id-missing", nil)
	require.NotEmpty(t, generated.Header.Get("X-Request-ID"),
		"every response carries a request id even for unknown routes")

	supplied := get(t, srv, "/probe/echo-request-id", map[string]string{"X-Request-ID": "abc-123"})
	require.Equal(t, "abc-123", supplied.Header.Get("X-Request-ID"))

	body, err := io.ReadAll(supplied.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "abc-123")
}
