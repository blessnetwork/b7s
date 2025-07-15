package traceable

import (
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "b7s.Store"
)

var defaultSpanOptions = []trace.SpanStartOption{
	trace.WithSpanKind(trace.SpanKindClient),
	trace.WithAttributes(semconv.DBSystemKey.String("pebble")),
}

func storeSpanOptions(opts ...trace.SpanStartOption) []trace.SpanStartOption {
	return append(defaultSpanOptions, opts...)
}
