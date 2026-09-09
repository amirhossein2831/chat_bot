package anthropic

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
	messagesPath = "/v1/messages"
	apiVersion   = "2023-06-01"

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

func WithModels(models []provider.ModelInfo) Option {
	return func(c *Client) { c.models = models }
}

func New(baseURL string, opts ...Option) *Client {
	c := &Client{
		http:    http.DefaultClient,
		baseURL: strings.TrimRight(baseURL, "/"),
		name:    "anthropic",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Client) Name() string { return c.name }

func (c *Client) Models() []provider.ModelInfo { return c.models }

// Health sends a one-token request because this API has no models endpoint to
// probe cheaply.
func (c *Client) Health(ctx context.Context) error {
	body, err := json.Marshal(chatRequest{
		Model:     c.probeModel(),
		Messages:  []chatMessage{{Role: string(provider.RoleUser), Content: "ping"}},
		MaxTokens: 1,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+messagesPath, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("%s health: status %d", c.name, resp.StatusCode)
	}
	return nil
}

func (c *Client) Stream(ctx context.Context, req provider.ChatRequest) (<-chan provider.Delta, error) {
	body, err := json.Marshal(toWire(req))
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+messagesPath, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	c.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")

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

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", apiVersion)
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}
}

func (c *Client) probeModel() string {
	if len(c.models) > 0 {
		return c.models[0].ID
	}
	return "claude-sonnet-5"
}
