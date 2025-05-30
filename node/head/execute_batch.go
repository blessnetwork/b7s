package head

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/request"
)

func (h *HeadNode) processExecuteBatch(ctx context.Context, from peer.ID, req request.ExecuteBatch) error {

	requestID := newRequestID()

	log := h.Log().With().
		Stringer("peer", from).
		Str("request", requestID).
		Str("function", req.Template.FunctionID).Logger()

	h.executeBatch(ctx, requestID, req)

}

func (h *HeadNode) executeBatch(ctx context.Context, requestID string, req request.ExecuteBatch) {

	// TODO: Metrics and tracing

	// Template request plus all others
	size := 1 + len(req.Arguments)

	log := h.Log().With().
		Str("request", requestID).
		Str("function", req.Template.FunctionID).
		Int("batch_size", size).
		Logger()

	peers, err := h.executeRollCall(ctx, requestID, req., 0)
	if err != nil {
		return fmt.Errorf("could not execute roll call: %w", err)
	}

}


