package worker

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/request"
)

func (w *Worker) processWorkOrderBatch(ctx context.Context, from peer.ID, req request.WorkOrderBatch) error {

}
