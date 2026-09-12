package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
)

const (
	chatPath    = "/v1/chat/completions"
	modelsPath  = "/v1/models"
	runningPath = "/api/ps"

	channelBuffer = 32
)

type Client struct {
	http     *http.Client
	baseURL  string
	apiKey   string
	name     string
	discover bool
	models   []provider.ModelInfo
}

type Option func(*Client)

func WithAPIKey(key string) Option { return func(c *Client) { c.apiKey = key } }

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.http = client }
}

func WithName(name string) Option { return func(c *Client) { c.name = name } }

// WithRuntimeDiscovery is for a self-hosted runtime, which knows how much
// context it loaded. A hosted provider is not asked, because its window is a
// published figure rather than a property of a machine.
func WithRuntimeDiscovery() Option { return func(c *Client) { c.discover = true } }

func WithModels(models []provider.ModelInfo) Option {
	return func(c *Client) { c.models = models }
}

// New serves both OpenAI and self-hosted vLLM: vLLM speaks the same API, so
// the only difference is the base URL and the absence of a key.
func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		http:    http.DefaultClient,
		baseURL: strings.TrimRight(baseURL, "/"),
		name:    "openai",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Name() string { return c.name }

func (c *Client) Models() []provider.ModelInfo { return c.models }

func (c *Client) Health(ctx context.Context) error {
	_, err := c.List(ctx)
	return err
}

func (c *Client) List(ctx context.Context) ([]string, error) {
	payload, err := c.list(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(payload.Data))
	for _, m := range payload.Data {
		ids = append(ids, m.ID)
	}
	return ids, nil
}

func (c *Client) list(ctx context.Context) (modelListResponse, error) {
	var payload modelListResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+modelsPath, nil)
	if err != nil {
		return payload, err
	}
	c.authorize(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return payload, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return payload, fmt.Errorf("%s health: status %d", c.name, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return payload, fmt.Errorf("%s models: %w", c.name, err)
	}
	return payload, nil
}

// ContextWindows reports the window each served model was actually loaded
// with. vLLM answers on the models endpoint; Ollama answers on its own, and
// its number is the one that matters because a prompt longer than the loaded
// context is silently cut rather than refused.
func (c *Client) ContextWindows(ctx context.Context) (map[string]int, error) {
	windows := map[string]int{}

	listed, err := c.list(ctx)
	if err != nil {
		return nil, err
	}
	for _, model := range listed.Data {
		if model.MaxModelLen > 0 {
			windows[model.ID] = model.MaxModelLen
		}
	}
	if len(windows) > 0 || !c.discover {
		return windows, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+runningPath, nil)
	if err != nil {
		return windows, nil
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return windows, nil
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return windows, nil
	}

	var running runningResponse
	if err := json.NewDecoder(resp.Body).Decode(&running); err != nil {
		return windows, nil
	}
	for _, model := range running.Models {
		if model.ContextLength > 0 {
			windows[model.Name] = model.ContextLength
		}
	}
	return windows, nil
}

func (c *Client) Stream(ctx context.Context, req provider.ChatRequest) (<-chan provider.Delta, error) {
	body, err := json.Marshal(toWire(req))
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+chatPath, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	c.authorize(httpReq)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		defer func() { _ = resp.Body.Close() }()
		return nil, fmt.Errorf("%s upstream: status %d", c.name, resp.StatusCode)
	}

	out := make(chan provider.Delta, channelBuffer)
	go func() {
		defer close(out)
		defer func() { _ = resp.Body.Close() }()
		emitStream(ctx, resp.Body, out)
	}()

	return out, nil
}

func (c *Client) authorize(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}
