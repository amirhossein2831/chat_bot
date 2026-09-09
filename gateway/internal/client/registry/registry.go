package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

const probeInterval = 30 * time.Second

type Registry struct {
	byModel       map[string]provider.Provider
	providerNames map[string]string
	models        []provider.ModelInfo
	providers     map[string]provider.Provider
	log           *zap.Logger

	metrics Availability

	mu        sync.RWMutex
	available map[string]bool
}

// Availability is satisfied by the metrics collector. It is an interface so
// the registry does not depend on the observability package.
type Availability interface {
	SetProviderAvailable(provider string, up bool)
}

func New(log *zap.Logger, metrics Availability, providers map[string]provider.Provider, models []provider.ModelInfo) *Registry {
	byModel := make(map[string]provider.Provider, len(models))
	providerNames := make(map[string]string, len(models))
	available := make(map[string]bool, len(providers))

	for _, m := range models {
		providerNames[m.ID] = m.Provider
		if p, ok := providers[m.Provider]; ok {
			byModel[m.ID] = p
		}
	}
	for name := range providers {
		available[name] = true
	}

	return &Registry{
		byModel:       byModel,
		providerNames: providerNames,
		models:        models,
		providers:     providers,
		log:           log,
		metrics:       metrics,
		available:     available,
	}
}

// Models reports live availability so the picker can grey out a tier before a
// user presses send, rather than failing at dispatch.
func (r *Registry) Models() []provider.ModelInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]provider.ModelInfo, 0, len(r.models))
	for _, m := range r.models {
		m.Available = r.available[m.Provider]
		out = append(out, m)
	}
	return out
}

func (r *Registry) Stream(ctx context.Context, req provider.ChatRequest) (<-chan provider.Delta, error) {
	target, err := r.route(req.Model)
	if err != nil {
		return nil, err
	}
	return target.Stream(ctx, req)
}

func (r *Registry) Health(context.Context) error { return nil }

func (r *Registry) ContextWindow(modelID string) int {
	for _, m := range r.models {
		if m.ID == modelID {
			return m.ContextWindow
		}
	}
	return 0
}

func (r *Registry) route(modelID string) (provider.Provider, error) {
	target, ok := r.byModel[modelID]
	if !ok {
		return nil, fmt.Errorf("%w: unknown model %q", constants.ErrModelUnavailable, modelID)
	}

	name := r.providerNames[modelID]
	r.mu.RLock()
	up := r.available[name]
	r.mu.RUnlock()

	if !up {
		return nil, fmt.Errorf("%w: %s is unreachable", constants.ErrModelUnavailable, name)
	}
	return target, nil
}

// Probe checks every provider concurrently. A local provider failing must not
// affect a frontier one, and vice versa, which is what keeps the free tier
// serving through a relay outage.
func (r *Registry) Probe(ctx context.Context) {
	var wg sync.WaitGroup
	results := make(map[string]bool, len(r.providers))
	var mu sync.Mutex

	for name, p := range r.providers {
		wg.Add(1)
		go func(name string, p provider.Provider) {
			defer wg.Done()

			probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			err := p.Health(probeCtx)
			mu.Lock()
			results[name] = err == nil
			mu.Unlock()

			if err != nil {
				r.log.Warn("provider unavailable", zap.String("provider", name), zap.Error(err))
			}
		}(name, p)
	}
	wg.Wait()

	r.mu.Lock()
	for name, up := range results {
		if r.metrics != nil {
			r.metrics.SetProviderAvailable(name, up)
		}
		if r.available[name] != up {
			r.log.Info("provider availability changed",
				zap.String("provider", name), zap.Bool("available", up))
		}
		r.available[name] = up
	}
	r.mu.Unlock()
}

func (r *Registry) Run(ctx context.Context) {
	ticker := time.NewTicker(probeInterval)
	defer ticker.Stop()

	r.Probe(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.Probe(ctx)
		}
	}
}
