package telemetry

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/zerologr"
	"github.com/hashicorp/go-metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
)

func InitializeTracing(ctx context.Context, log zerolog.Logger, opts ...TraceOption) (ShutdownFunc, error) {

	cfg := DefaultTraceConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	resource, err := CreateResource(ctx, cfg.ID, cfg.Role)
	if err != nil {
		return nil, fmt.Errorf("could not initialize otel resource: %w", err)
	}

	// Setup general otel stuff.
	setupOtel(log)

	// Setup tracing.
	exporters, err := createTraceExporters(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create trace exporters: %w", err)
	}

	if cfg.ExporterBatchTimeout == 0 {
		log.Warn().Msg("trace exporter batch timeout is disabled")
	}

	tp := CreateTracerProvider(resource, cfg.ExporterBatchTimeout, exporters...)
	otel.SetTracerProvider(tp)

	// From here on down, we have components that need shutdown.
	// Potential other shutdown functions should be appended to this slice.
	shutdownFuncs := []ShutdownFunc{tp.Shutdown}

	return shutdownAll(shutdownFuncs), nil
}

func InitializeMetrics(opts ...MetricsOption) (*metrics.Metrics, error) {

	cfg := DefaultMetricsConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	registerer := prometheus.DefaultRegisterer

	sink, err := CreateMetricSink(registerer, cfg)
	if err != nil {
		return nil, fmt.Errorf("could not create prometheus sink: %w", err)
	}

	m, err := CreateMetrics(sink, cfg.Global)
	if err != nil {
		return nil, fmt.Errorf("could not create prometheus metrics: %w", err)
	}

	return m, nil
}

func shutdownAll(funcs []ShutdownFunc) ShutdownFunc {
	return func(ctx context.Context) error {
		var err error
		for _, fn := range funcs {
			err = errors.Join(err, fn(ctx))
		}
		return err
	}
}

// Setup general otel stuff like logging and error handling. We will just log telemetry errors and do nothing more.
func setupOtel(log zerolog.Logger) {

	otel.SetTextMapPropagator(CreatePropagator())
	otel.SetLogger(zerologr.New(&log))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(
		func(err error) { log.Error().Err(err).Msg("telemetry error") }),
	)
}
