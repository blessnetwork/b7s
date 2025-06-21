package head

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/node"
)

// processMessage will determine which message was received and how to process it.
func (h *HeadNode) process(ctx context.Context, from peer.ID, msg string, payload []byte) error {

	switch msg {
	case bls.MessageHealthCheck:
		return node.HandleMessage(ctx, from, payload, h.processHealthCheck)
	case bls.MessageInstallFunctionResponse:
		return node.HandleMessage(ctx, from, payload, h.processInstallFunctionResponse)
	case bls.MessageExecute:
		return node.HandleMessage(ctx, from, payload, h.processExecute)
	case bls.MessageRollCallResponse:
		return node.HandleMessage(ctx, from, payload, h.processRollCallResponse)
	case bls.MessageWorkOrderResponse:
		return node.HandleMessage(ctx, from, payload, h.processWorkOrderResponse)
	case bls.MessageFormClusterResponse:
		return node.HandleMessage(ctx, from, payload, h.processFormClusterResponse)
	case bls.MessageExecuteBatch:
		return node.HandleMessage(ctx, from, payload, h.processExecuteBatch)
	case bls.MessageWorkOrderBatchResponse:
		return node.HandleMessage(ctx, from, payload, h.processWorkOrderBatchResponse)
	}

	return fmt.Errorf("unsupported message: %s", msg)
}
