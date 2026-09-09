package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Time to first token is the primary metric: it is what a user perceives as
// speed, and it is the first thing to degrade when a relay is struggling.
type Metrics struct {
	timeToFirstToken *prometheus.HistogramVec
	streamDuration   *prometheus.HistogramVec
	streamTokens     *prometheus.CounterVec
	upstreamErrors   *prometheus.CounterVec
	providerUp       *prometheus.GaugeVec
	quotaRejections  *prometheus.CounterVec
	usageCost        *prometheus.CounterVec
	requests         *prometheus.HistogramVec
}

func NewMetrics(registerer prometheus.Registerer) *Metrics {
	m := &Metrics{
		timeToFirstToken: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "chat_time_to_first_token_seconds",
			Help:    "Delay between dispatching a generation and its first token.",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 4, 8, 16, 32},
		}, []string{"model"}),

		streamDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "chat_stream_duration_seconds",
			Help:    "Wall time of a whole generation.",
			Buckets: []float64{0.5, 1, 2, 5, 10, 30, 60, 120, 300},
		}, []string{"model", "outcome"}),

		streamTokens: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "chat_tokens_total",
			Help: "Tokens counted from settled generations.",
		}, []string{"model", "direction"}),

		upstreamErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "provider_upstream_errors_total",
			Help: "Errors returned by a provider mid-stream.",
		}, []string{"model"}),

		providerUp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "provider_available",
			Help: "Whether a provider answered its last health probe.",
		}, []string{"provider"}),

		quotaRejections: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "quota_rejections_total",
			Help: "Requests refused before dispatch.",
		}, []string{"reason"}),

		usageCost: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "usage_cost_usd_micros_total",
			Help: "Cost of settled generations, in USD micro-cents.",
		}, []string{"model"}),

		requests: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route", "status"}),
	}

	registerer.MustRegister(
		m.timeToFirstToken, m.streamDuration, m.streamTokens,
		m.upstreamErrors, m.providerUp, m.quotaRejections, m.usageCost, m.requests,
	)
	return m
}

func (m *Metrics) ObserveFirstToken(model string, d time.Duration) {
	m.timeToFirstToken.WithLabelValues(model).Observe(d.Seconds())
}

func (m *Metrics) ObserveStream(model, outcome string, d time.Duration) {
	m.streamDuration.WithLabelValues(model, outcome).Observe(d.Seconds())
}

func (m *Metrics) ObserveUsage(model string, in, out int, costMicros int64) {
	m.streamTokens.WithLabelValues(model, "in").Add(float64(in))
	m.streamTokens.WithLabelValues(model, "out").Add(float64(out))
	m.usageCost.WithLabelValues(model).Add(float64(costMicros))
}

func (m *Metrics) IncUpstreamError(model string) {
	m.upstreamErrors.WithLabelValues(model).Inc()
}

func (m *Metrics) SetProviderAvailable(provider string, up bool) {
	value := 0.0
	if up {
		value = 1
	}
	m.providerUp.WithLabelValues(provider).Set(value)
}

func (m *Metrics) IncQuotaRejection(reason string) {
	m.quotaRejections.WithLabelValues(reason).Inc()
}

func (m *Metrics) ObserveRequest(method, route, status string, d time.Duration) {
	m.requests.WithLabelValues(method, route, status).Observe(d.Seconds())
}
