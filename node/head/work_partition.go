package head

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/request"
)

// partitionWorkBatch takes a work batch (which can contain a large number of variants of the same execution)
// and splits them among a number of workers.
//
// In the future, we may have different criteria for what gets assigned to each peer. Right now we do round robin.
func partitionWorkBatch(peers []peer.ID, requestID string, req request.ExecuteBatch) map[peer.ID]*request.WorkOrderBatch {

	variants := req.Arguments

	// TODO: Do this in one go, not two maps.

	// Assign arguments to a list of peers in a round robin fashion
	n := len(peers)
	a := make(map[peer.ID][][]string)
	for i, args := range variants {
		target := peers[i%n]

		a[target] = append(a[target], args)
	}

	assignments := make(map[peer.ID]*request.WorkOrderBatch)
	for _, peer := range peers {

		strandID := newStrandID(requestID)
		assignments[peer] = req.WorkOrderBatch(requestID, strandID, a[peer]...)
	}

	return assignments
}

func newStrandID(requestID string) string {
	return fmt.Sprintf("%v:%v", requestID, newRequestID())
}
