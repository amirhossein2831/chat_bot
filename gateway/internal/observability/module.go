package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
)

var Module = fx.Module("observability",
	fx.Provide(
		func() *prometheus.Registry { return prometheus.NewRegistry() },
		func(r *prometheus.Registry) prometheus.Gatherer { return r },
		func(r *prometheus.Registry) *Metrics { return NewMetrics(r) },
	),
)
