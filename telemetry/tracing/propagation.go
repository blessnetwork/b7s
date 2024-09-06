package tracing

import (
	"context"
	"encoding/json"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type TraceInfo struct {
	Carrier propagation.MapCarrier
}

// Empty returns true if the TraceInfo structure contains any tracing information.
func (t TraceInfo) Empty() bool {
	return len(t.Carrier.Keys()) == 0
}

// GetTraceInfo extracts tracing information from the context.
func GetTraceInfo(ctx context.Context) TraceInfo {
	return GetTraceInfoWithPropagator(ctx, otel.GetTextMapPropagator())
}

func GetTraceInfoWithPropagator(ctx context.Context, propagator propagation.TextMapPropagator) TraceInfo {

	carrier := propagation.MapCarrier{}
	propagator.Inject(ctx, carrier)

	return TraceInfo{Carrier: carrier}
}

// TraceContextFromMessage will try to extract TraceInfo from the JSON message.
func TraceContextFromMessage(ctx context.Context, payload []byte) (context.Context, error) {

	var traceInfo TraceInfo
	err := json.Unmarshal(payload, &traceInfo)
	if err != nil {
		return ctx, fmt.Errorf("could not extract trace info from context: %w", err)
	}

	return TraceContext(ctx, traceInfo), nil
}

// TraceContext injects the trace information into passed context.
func TraceContext(ctx context.Context, t TraceInfo) context.Context {
	return TraceContextWithPropagator(ctx, otel.GetTextMapPropagator(), t)
}

func TraceContextWithPropagator(ctx context.Context, propagator propagation.TextMapPropagator, t TraceInfo) context.Context {
	return propagator.Extract(ctx, t.Carrier)

}
