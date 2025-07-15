package telemetry

import (
	"testing"

	"github.com/hashicorp/go-metrics/prometheus"
	"github.com/stretchr/testify/require"
)

func TestMetricsConfig_MetricCounters(t *testing.T) {

	var counters = []prometheus.CounterDefinition{
		{
			Name: []string{"random", "counter", "value", "1"},
			Help: "Dummy counter description",
		},
		{
			Name: []string{"generic", "counter", "value", "2"},
			Help: "Dummy counter description",
		},
	}

	var cfg MetricsConfig
	WithCounters(counters)(&cfg)
	require.Equal(t, counters, cfg.Counters)
}

func TestMetricsConfig_MetricSummaries(t *testing.T) {

	var summary = []prometheus.SummaryDefinition{
		{
			Name: []string{"random", "summary", "value", "1"},
			Help: "Dummy summary description #1",
		},
		{
			Name: []string{"generic", "summary", "value", "2"},
			Help: "Dummy summary description #2",
		},
	}

	var cfg MetricsConfig
	WithSummaries(summary)(&cfg)
	require.Equal(t, summary, cfg.Summaries)
}

func TestMetricsConfig_MetricGauges(t *testing.T) {

	var gauges = []prometheus.GaugeDefinition{
		{
			Name: []string{"random", "gauges", "value", "1"},
			Help: "Dummy gauges description",
		},
		{
			Name: []string{"generic", "gauges", "value", "2"},
			Help: "Dummy gauges description",
		},
	}

	var cfg MetricsConfig
	WithGauges(gauges)(&cfg)
	require.Equal(t, gauges, cfg.Gauges)
}
