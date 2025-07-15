package telemetry

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/go-metrics"
	mp "github.com/hashicorp/go-metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func CreateMetricSink(registerer prometheus.Registerer, cfg MetricsConfig) (*mp.PrometheusSink, error) {

	opts := mp.PrometheusOpts{
		Registerer:         registerer,
		CounterDefinitions: formatCounters(cfg.Counters),
		SummaryDefinitions: formatSummaries(cfg.Summaries),
		GaugeDefinitions:   formatGauges(cfg.Gauges),
	}

	return mp.NewPrometheusSinkFrom(opts)
}

func CreateMetrics(sink *mp.PrometheusSink, global bool) (*metrics.Metrics, error) {

	mcfg := &metrics.Config{
		ServiceName:          "b7s",
		FilterDefault:        true,
		EnableRuntimeMetrics: true,
		TimerGranularity:     time.Millisecond,
	}

	var (
		m   *metrics.Metrics
		err error
	)

	if global {
		m, err = metrics.NewGlobal(mcfg, sink)
	} else {
		m, err = metrics.New(mcfg, sink)
	}
	if err != nil {
		return nil, fmt.Errorf("could not create new metrics instance: %w", err)
	}

	return m, nil
}

// GetMetricsHTTPHandler returns an HTTP handler for the default prometheus registerer and gatherer.
func GetMetricsHTTPHandler() http.Handler {

	opts := promhttp.HandlerOpts{
		Registry: prometheus.DefaultRegisterer,
	}

	return promhttp.HandlerFor(prometheus.DefaultGatherer, opts)
}

func formatCounters(counters []mp.CounterDefinition) []mp.CounterDefinition {

	prefixed := make([]mp.CounterDefinition, len(counters))

	for i := 0; i < len(counters); i++ {
		c := counters[i]
		c.Name = append([]string{metricPrefix}, c.Name...)
		prefixed[i] = c
	}

	return prefixed
}

func formatSummaries(summaries []mp.SummaryDefinition) []mp.SummaryDefinition {

	prefixed := make([]mp.SummaryDefinition, len(summaries))

	for i := 0; i < len(summaries); i++ {
		s := summaries[i]
		s.Name = append([]string{metricPrefix}, s.Name...)
		prefixed[i] = s
	}

	return prefixed
}

func formatGauges(gauges []mp.GaugeDefinition) []mp.GaugeDefinition {

	// Right now we have a single gauge - node info.
	// gauges := node.Gauges
	prefixed := make([]mp.GaugeDefinition, len(gauges))

	for i := 0; i < len(gauges); i++ {
		g := gauges[i]
		g.Name = append([]string{metricPrefix}, g.Name...)
		prefixed[i] = g
	}

	return prefixed
}
