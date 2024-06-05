package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/blocklessnetwork/b7s/models/blockless"
	"github.com/blocklessnetwork/b7s/models/execute"
	"github.com/blocklessnetwork/b7s/telemetry/b7ssemconv"
)

type Tracer struct {
	trace.Tracer
}

func NewTracer(name string, opts ...trace.TracerOption) *Tracer {

	return &Tracer{
		Tracer: otel.Tracer(name, opts...),
	}
}

func (t *Tracer) WithSpanFromContext(ctx context.Context, spanName string, f func() error, opts ...trace.SpanStartOption) error {

	_, span := t.Start(ctx, spanName, opts...)
	defer span.End()

	err := f()
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

func SpanAttributes(attributes []attribute.KeyValue) []trace.SpanStartOption {
	return []trace.SpanStartOption{(trace.WithAttributes(attributes...))}

}

func ExecutionAttributes(requestID string, req execute.Request) []attribute.KeyValue {
	return []attribute.KeyValue{
		b7ssemconv.FunctionCID.String(req.FunctionID),
		b7ssemconv.FunctionMethod.String(req.Method),
		b7ssemconv.ExecutionNodeCount.Int(req.Config.NodeCount),
		b7ssemconv.ExecutionConsensus.String(req.Config.ConsensusAlgorithm),
		b7ssemconv.ExecutionRequestID.String(requestID),
	}
}

func PeerAttributes(peer blockless.Peer) []attribute.KeyValue {
	return []attribute.KeyValue{
		b7ssemconv.PeerID.String(peer.ID.String()),
		b7ssemconv.PeerMultiaddr.String(peer.MultiAddr),
	}
}
