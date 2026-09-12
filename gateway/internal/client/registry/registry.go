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
	modelUp   map[string]bool
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
	modelUp := make(map[string]bool, len(models))

	for _, m := range models {
		providerNames[m.ID] = m.Provider
		if p, ok := providers[m.Provider]; ok {
			byModel[m.ID] = p
		}
		modelUp[m.ID] = true
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
		modelUp:       modelUp,
	}
}

// Models reports live availability so the picker can grey out a tier before a
// user presses send, rather than failing at dispatch.
func (r *Registry) Models() []provider.ModelInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]provider.ModelInfo, 0, len(r.models))
	for _, m := range r.models {
		m.Available = r.modelUp[m.ID]
		out = append(out, m)
	}
	return out
}

func (r *Registry) Stream(ctx context.Context, req provider.ChatRequest) (<-chan provider.Delta, error) {
	target, err := r.route(req.Model)
	if err != nil {
		return nil, err
	}

	deltas, err := target.Stream(ctx, req)
	if err != nil {
		// A real request failing is better evidence than a probe that has not
		// run yet: without this the model stays "available" for up to the probe
		// interval while every request fails.
		r.markUnavailable(req.Model)
		return nil, err
	}
	return deltas, nil
}

// markUnavailable narrows to the single model that failed. One model missing
// from a local runtime says nothing about its siblings on the same provider.
func (r *Registry) markUnavailable(modelID string) {
	if modelID == "" {
		return
	}

	r.mu.Lock()
	changed := r.modelUp[modelID]
	r.modelUp[modelID] = false
	r.mu.Unlock()

	if changed {
		r.log.Warn("model marked unavailable after a failed request",
			zap.String("model", modelID))
	}
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
	providerUp := r.available[name]
	up := r.modelUp[modelID]
	r.mu.RUnlock()

	if !providerUp {
		return nil, fmt.Errorf("%w: %s is unreachable", constants.ErrModelUnavailable, name)
	}
	if !up {
		return nil, fmt.Errorf("%w: %s is not loaded on %s", constants.ErrModelUnavailable, modelID, name)
	}
	return target, nil
}

// Probe checks every provider concurrently. A local provider failing must not
// affect a frontier one, and vice versa, which is what keeps the free tier
// serving through a relay outage.
func (r *Registry) Probe(ctx context.Context) {
	var wg sync.WaitGroup
	results := make(map[string]bool, len(r.providers))
	served := make(map[string]map[string]bool, len(r.providers))
	var mu sync.Mutex

	for name, p := range r.providers {
		wg.Add(1)
		go func(name string, p provider.Provider) {
			defer wg.Done()

			probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			up, ids := r.inspect(probeCtx, name, p)

			mu.Lock()
			results[name] = up
			served[name] = ids
			mu.Unlock()
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

	for _, m := range r.models {
		up, probed := results[m.Provider]
		if !probed {
			continue
		}

		ids := served[m.Provider]
		loaded := up && (ids == nil || ids[m.ID])
		if r.modelUp[m.ID] != loaded {
			r.log.Info("model availability changed",
				zap.String("model", m.ID), zap.Bool("available", loaded))
		}
		r.modelUp[m.ID] = loaded
	}
	r.mu.Unlock()
}

// inspect returns whether the provider answered and, when it can say, the set
// of models it is serving. A nil set means the provider does not report one,
// so every model the catalogue assigns to it is taken on trust.
func (r *Registry) inspect(ctx context.Context, name string, p provider.Provider) (bool, map[string]bool) {
	lister, ok := p.(provider.Lister)
	if !ok {
		if err := p.Health(ctx); err != nil {
			r.log.Warn("provider unavailable", zap.String("provider", name), zap.Error(err))
			return false, nil
		}
		return true, nil
	}

	ids, err := lister.List(ctx)
	if err != nil {
		r.log.Warn("provider unavailable", zap.String("provider", name), zap.Error(err))
		return false, nil
	}

	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return true, set
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
