package head

import (
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/request"
)

// partitionWorkBatch takes a work batch (which can contain a large number of variants of the same execution)
// and splits them among a number of workers.
//
// In the future, we may have different criteria for what gets assigned to each peer. Right now we do round robin.
func partitionWorkBatch(peers []peer.ID, requestID string, req request.ExecuteBatch) map[peer.ID]*request.WorkOrderBatch {

	n := len(peers)
	assignments := make(map[peer.ID]*request.WorkOrderBatch)

	if n == 0 {
		return assignments
	}

	variants := req.Arguments

	// TODO: Do this in one go, not two maps.

	// Assign arguments to a list of peers in a round robin fashion
	a := make(map[peer.ID][][]string)
	for i, args := range variants {
		target := peers[i%n]
		a[target] = append(a[target], args)
	}

	for _, peer := range peers {
		chunkID := newChunkID()
		assignments[peer] = req.WorkOrderBatch(requestID, chunkID, a[peer]...)
	}

	return assignments
}

func newChunkID() string {
	return newUUID()
}
