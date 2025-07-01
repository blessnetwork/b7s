package node

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-metrics"
	"github.com/libp2p/go-libp2p/core/peer"
	otelcodes "go.opentelemetry.io/otel/codes"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/telemetry/tracing"
)

func (c *core) processMessage(ctx context.Context, from peer.ID, payload []byte, pipeline Pipeline, process func(context.Context, peer.ID, string, []byte) error) (procError error) {

	// Determine message type.
	typ, err := getMessageType(payload)
	if err != nil {
		return fmt.Errorf("could not unpack message: %w", err)
	}

	log := c.log.With().Stringer("peer", from).Str("type", typ).Stringer("pipeline", pipeline).Logger()

	c.metrics.IncrCounterWithLabels(messagesProcessedMetric, 1, []metrics.Label{{Name: "type", Value: typ}})
	defer func() {
		switch procError {
		case nil:
			c.metrics.IncrCounterWithLabels(messagesProcessedOkMetric, 1, []metrics.Label{{Name: "type", Value: typ}})
		default:
			c.metrics.IncrCounterWithLabels(messagesProcessedErrMetric, 1, []metrics.Label{{Name: "type", Value: typ}})
		}
	}()

	ctx, err = tracing.TraceContextFromMessage(ctx, payload)
	if err != nil {
		c.log.Error().Err(err).Msg("could not get trace context from message")
	}

	ctx, span := c.tracer.Start(ctx, msgProcessSpanName(typ), msgProcessSpanOpts(from, typ, pipeline)...)
	defer span.End()
	// NOTE: This function checks the named return error value in order to set the span status accordingly.
	defer func() {
		if procError == nil {
			span.SetStatus(otelcodes.Ok, spanStatusOK)
			return
		}

		if allowErrorLeakToTelemetry {
			span.SetStatus(otelcodes.Error, procError.Error())
			return
		}

		span.SetStatus(otelcodes.Error, spanStatusErr)
	}()

	if !correctPipeline(typ, pipeline) {
		log.Warn().Msg("message not allowed on pipeline")
		return nil
	}

	return process(ctx, from, typ, payload)
}

func HandleMessage[T bls.Message](ctx context.Context, from peer.ID, payload []byte, processFunc func(ctx context.Context, from peer.ID, msg T) error) error {

	var msg T
	err := json.Unmarshal(payload, &msg)
	if err != nil {
		return fmt.Errorf("could not unmarshal message: %w", err)
	}

	// If the message provides a validation mechanism - use it.
	type validator interface {
		Valid() error
	}

	vmsg, ok := any(msg).(validator)
	if ok {
		err = vmsg.Valid()
		if err != nil {
			return fmt.Errorf("rejecting message that failed validation: %w", err)
		}
	}

	return processFunc(ctx, from, msg)
}

// getMessageType will return the `type` string field from the JSON payload.
func getMessageType(payload []byte) (string, error) {

	type baseMessage struct {
		Type string `json:"type,omitempty"`
	}
	var message baseMessage
	err := json.Unmarshal(payload, &message)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal message: %w", err)
	}

	return message.Type, nil
}
