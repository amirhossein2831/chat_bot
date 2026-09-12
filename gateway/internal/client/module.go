package client

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/anthropic"
	"github.com/amirhossein2831/chat_bot/gateway/internal/client/openai"
	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
	"github.com/amirhossein2831/chat_bot/gateway/internal/client/registry"
	"github.com/amirhossein2831/chat_bot/gateway/internal/client/relay"
	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/observability"
)

// streamTimeout bounds a whole generation. It is not a read timeout: a model
// can legitimately think for a long time before the first token.
const streamTimeout = 15 * time.Minute

var Module = fx.Module("client",
	fx.Provide(NewRegistry),
)

func NewRegistry(lc fx.Lifecycle, cfg *config.Config, metrics *observability.Metrics, log *zap.Logger) *registry.Registry {
	models := toModelInfo(cfg.Models)
	providers := buildProviders(cfg, models)

	reg := registry.New(log, metrics, providers, models)

	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go reg.Run(ctx)
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})

	return reg
}

func buildProviders(cfg *config.Config, models []provider.ModelInfo) map[string]provider.Provider {
	direct := &http.Client{Timeout: streamTimeout}

	// Frontier traffic goes through the relay when one is configured, so
	// provider keys stay outside Iran. Without endpoints it falls back to a
	// direct client, which is what local development uses.
	frontier := direct
	if len(cfg.Relay.Endpoints) > 0 {
		frontier = &http.Client{
			Timeout:   streamTimeout,
			Transport: relay.NewTransport(cfg.Relay.Endpoints, cfg.Relay.Credential, nil),
		}
	}

	providers := map[string]provider.Provider{}

	if cfg.Providers.Local.BaseURL != "" {
		providers["local"] = openai.New(cfg.Providers.Local.BaseURL,
			openai.WithName("local"),
			openai.WithRuntimeDiscovery(),
			openai.WithHTTPClient(direct),
			openai.WithModels(modelsFor(models, "local")))
	}
	if cfg.Providers.OpenAI.BaseURL != "" {
		providers["openai"] = openai.New(cfg.Providers.OpenAI.BaseURL,
			openai.WithName("openai"),
			openai.WithAPIKey(cfg.Providers.OpenAI.APIKey),
			openai.WithHTTPClient(frontier),
			openai.WithModels(modelsFor(models, "openai")))
	}
	if cfg.Providers.Anthropic.BaseURL != "" {
		providers["anthropic"] = anthropic.New(cfg.Providers.Anthropic.BaseURL,
			anthropic.WithAPIKey(cfg.Providers.Anthropic.APIKey),
			anthropic.WithHTTPClient(frontier),
			anthropic.WithModels(modelsFor(models, "anthropic")))
	}

	return providers
}

func toModelInfo(models []config.ModelConfig) []provider.ModelInfo {
	out := make([]provider.ModelInfo, 0, len(models))
	for _, m := range models {
		out = append(out, provider.ModelInfo{
			ID:            m.ID,
			Provider:      m.Provider,
			Tier:          provider.Tier(m.Tier),
			ContextWindow: m.ContextWindow,
		})
	}
	return out
}

func modelsFor(models []provider.ModelInfo, providerName string) []provider.ModelInfo {
	out := make([]provider.ModelInfo, 0, len(models))
	for _, m := range models {
		if m.Provider == providerName {
			out = append(out, m)
		}
	}
	return out
}
