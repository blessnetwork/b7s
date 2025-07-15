package telemetry_test

import (
	"strings"
	"testing"

	mp "github.com/hashicorp/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"

	"github.com/blessnetwork/b7s/telemetry"
)

func TestTelemetry_Metrics(t *testing.T) {

	var (
		counters = []mp.CounterDefinition{
			{Help: "dummy counter #1", Name: []string{"test", "counter", "definition", "1"}},
			{Help: "dummy counter #2", Name: []string{"test", "counter", "definition", "2"}},
		}

		summaries = []mp.SummaryDefinition{
			{Help: "dummy summary #1", Name: []string{"test", "summary", "definition", "1"}},
			{Help: "dummy summary #2", Name: []string{"test", "summary", "definition", "2"}},
		}

		gauges = []mp.GaugeDefinition{
			{Help: "dummy gauge #1", Name: []string{"test", "gauge", "definition", "1"}},
			{Help: "dummy gauge #2", Name: []string{"test", "gauge", "definition", "2"}},
		}
	)

	registry := prometheus.NewRegistry()

	cfg := telemetry.MetricsConfig{
		Counters:  counters,
		Summaries: summaries,
		Gauges:    gauges,
	}

	sink, err := telemetry.CreateMetricSink(registry, cfg)
	require.NoError(t, err)

	_, err = telemetry.CreateMetrics(sink, false)
	require.NoError(t, err)

	gathered, err := registry.Gather()
	require.NoError(t, err)

	// Create a map of gathered metrics.
	metrics := make(map[string]*dto.MetricFamily)
	for _, m := range gathered {
		metrics[*m.Name] = m
	}

	// Create a map of original metrics.
	original := createMetricMap(counters, summaries, gauges)

	for name, desc := range original {

		metric, ok := metrics["b7s_"+name]
		require.True(t, ok)
		require.Equal(t, desc, metric.GetHelp())

		switch name {
		case "test_counters_definition_1",
			"test_counters_definition_2":
			require.Equal(t, dto.MetricType_COUNTER, metric.GetType())

		case "test_summary_definition_1",
			"test_summary_definition_2":
			require.Equal(t, dto.MetricType_SUMMARY, metric.GetType())

		case "test_gauge_definition_1",
			"test_gauge_definition_2":
			require.Equal(t, dto.MetricType_GAUGE, metric.GetType())
		}
	}
}

func createMetricMap(counters []mp.CounterDefinition, summaries []mp.SummaryDefinition, gauges []mp.GaugeDefinition) map[string]string {

	out := make(map[string]string)
	for _, c := range counters {
		out[strings.Join(c.Name, "_")] = c.Help
	}
	for _, s := range summaries {
		out[strings.Join(s.Name, "_")] = s.Help
	}
	for _, g := range gauges {
		out[strings.Join(g.Name, "_")] = g.Help
	}

	return out
}
