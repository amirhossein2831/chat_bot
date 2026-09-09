package relay

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
)

// Transport sends a provider request through a relay hosted outside Iran and
// attaches the gateway's own credential. Provider keys are never held on this
// side; the relay injects them.
type Transport struct {
	endpoints  []string
	credential string
	base       http.RoundTripper
	next       atomic.Uint64
}

func NewTransport(endpoints []string, credential string, base http.RoundTripper) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}
	return &Transport{
		endpoints:  endpoints,
		credential: credential,
		base:       base,
	}
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.endpoints) == 0 {
		return nil, fmt.Errorf("relay: no endpoints configured")
	}

	start := int(t.next.Add(1)) % len(t.endpoints)
	var lastErr error

	// Round-robin with failover: try every relay before giving up, so one
	// blocked host costs latency rather than availability.
	for offset := range t.endpoints {
		endpoint := t.endpoints[(start+offset)%len(t.endpoints)]

		attempt, err := t.rewrite(req, endpoint)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := t.base.RoundTrip(attempt)
		if err == nil && resp.StatusCode < http.StatusInternalServerError {
			return resp, nil
		}
		if resp != nil {
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("relay %s: status %d", endpoint, resp.StatusCode)
		}
		if err != nil {
			lastErr = err
		}
	}

	return nil, fmt.Errorf("relay: all endpoints failed: %w", lastErr)
}

func (t *Transport) rewrite(req *http.Request, endpoint string) (*http.Request, error) {
	target, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	attempt := req.Clone(req.Context())
	attempt.URL.Scheme = target.Scheme
	attempt.URL.Host = target.Host
	attempt.URL.Path = strings.TrimRight(target.Path, "/") + req.URL.Path
	attempt.Host = target.Host
	attempt.Header.Set("Authorization", "Bearer "+t.credential)

	// A failed attempt has already consumed the body, so each retry needs a
	// fresh one. http.NewRequest sets GetBody for the readers the clients use.
	if req.GetBody != nil {
		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		attempt.Body = body
	}

	return attempt, nil
}
