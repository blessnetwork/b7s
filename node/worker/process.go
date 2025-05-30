package worker

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/node"
)

func (w *Worker) process(ctx context.Context, from peer.ID, msg string, payload []byte) error {

	switch msg {
	case bls.MessageHealthCheck:
		return node.HandleMessage(ctx, from, payload, w.processHealthCheck)
	case bls.MessageInstallFunction:
		return node.HandleMessage(ctx, from, payload, w.processInstallFunction)
	case bls.MessageRollCall:
		return node.HandleMessage(ctx, from, payload, w.processRollCall)
	case bls.MessageWorkOrder:
		return node.HandleMessage(ctx, from, payload, w.processWorkOrder)
	case bls.MessageFormCluster:
		return node.HandleMessage(ctx, from, payload, w.processFormCluster)
	case bls.MessageDisbandCluster:
		return node.HandleMessage(ctx, from, payload, w.processDisbandCluster)
	case bls.MessageWorkOrderBatch:
		return node.HandleMessage(ctx, from, payload, w.processWorkOrderBatch)
	}

	return fmt.Errorf("unsupported message: %s", msg)
}
