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
	chatPath   = "/v1/chat/completions"
	modelsPath = "/v1/models"

	channelBuffer = 32
)

type Client struct {
	http    *http.Client
	baseURL string
	apiKey  string
	name    string
	models  []provider.ModelInfo
}

type Option func(*Client)

func WithAPIKey(key string) Option { return func(c *Client) { c.apiKey = key } }

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) { c.http = client }
}

func WithName(name string) Option { return func(c *Client) { c.name = name } }

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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+modelsPath, nil)
	if err != nil {
		return err
	}
	c.authorize(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("%s health: status %d", c.name, resp.StatusCode)
	}
	return nil
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
