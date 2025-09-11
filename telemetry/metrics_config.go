package telemetry

import (
	"github.com/hashicorp/go-metrics/prometheus"
)

var DefaultMetricsConfig = MetricsConfig{
	Global: true,
}

type MetricsConfig struct {
	Global    bool
	Counters  []prometheus.CounterDefinition
	Summaries []prometheus.SummaryDefinition
	Gauges    []prometheus.GaugeDefinition
}

type MetricsOption func(*MetricsConfig)

func WithCounters(counters []prometheus.CounterDefinition) MetricsOption {
	return func(cfg *MetricsConfig) {
		cfg.Counters = counters
	}
}

func WithSummaries(summaries []prometheus.SummaryDefinition) MetricsOption {
	return func(cfg *MetricsConfig) {
		cfg.Summaries = summaries
	}
}

func WithGauges(gauges []prometheus.GaugeDefinition) MetricsOption {
	return func(cfg *MetricsConfig) {
		cfg.Gauges = gauges
	}
}
