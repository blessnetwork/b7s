package head

import (
	"context"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/models/response"
)

type ExecutionBatchAssignments map[peer.ID]*request.WorkOrderBatch

// NOTE: Batch execution is a special case of an execution. Instead of issuing an execution request/work order
// to nodes to execute a single thing (naturally one function with a set of arguments) we want to
// task many nodes to execute one function with a large number of arguments.
//
// Since it's still unclear if this is here to stay and/or if it becomes a canon method of execution in b7s,
// in order to not alter the default way of execution, this use case is handled separately, even though there are big overlaps.
func (h *HeadNode) processExecuteBatch(ctx context.Context, from peer.ID, req request.ExecuteBatch) error {

	requestID := newRequestID()

	log := h.Log().With().
		Stringer("peer", from).
		Str("request", requestID).
		Str("function", req.Template.FunctionID).
		Int("size", len(req.Arguments)).Logger()

	log.Info().Msg("received a batch request")

	results, err := h.executeBatch(ctx, requestID, req)
	if err != nil {
		return fmt.Errorf("could not execute batch request: %w", err)
	}

	log.Info().Any("results", results).Msg("received batch responses")

	// TODO: Add actual status code.
	res := req.Response(codes.OK, requestID).WithResults(results)

	err = h.Send(ctx, from, res)
	if err != nil {
		return fmt.Errorf("could not send batch response: %w", err)
	}

	return nil
}

type batchResults map[string]response.NodeStrandResults

func (h *HeadNode) executeBatch(ctx context.Context, requestID string, req request.ExecuteBatch) (batchResults, error) {

	// TODO: Metrics and tracing

	// Template request plus all others
	size := len(req.Arguments)

	log := h.Log().With().
		Str("request", requestID).
		Str("function", req.Template.FunctionID).
		Int("batch_size", size).
		Logger()

	log.Info().Msg("processing batch execution request")

	// Phase 1. - Issue roll call to nodes.

	rc := rollCallRequest(req.Template.FunctionID, requestID, 0, req.Template.Config.Attributes)

	rctx, cancel := context.WithTimeout(ctx, h.cfg.ExecutionTimeout)
	defer cancel()

	// node count is -1 - we want all the nodes that want to work.
	peers, err := h.executeRollCall(rctx, rc, req.Topic, req.Template.Config.NodeCount)
	if err != nil {
		return nil, fmt.Errorf("could not execute roll call: %w", err)
	}

	log.Debug().
		Strs("peers", bls.PeerIDsToStr(peers)).
		Msg("peers reported for work")

	assignments := partitionWorkBatch(peers, requestID, req)

	// TODO: Rethink, useful but ugly.
	logAssignments(&log, assignments)

	err = h.sendBatch(ctx, assignments)
	if err != nil {

		var sendErr *batchSendError
		if errors.As(err, &sendErr) {
			// TODO: Handle partial failures by retrying part of the batch that failed.
			log.Warn().
				Strs("peers", bls.PeerIDsToStr(sendErr.Targets())).
				Msg("partial failure to send batch requst")
		}

		return nil, fmt.Errorf("could not send work order batch: %w", err)
	}

	// TODO: Handle errors - reintroduce to the pool.

	// Wait for results.

	assignedWorkers := mapKeys(assignments)

	waitctx, cancel := context.WithTimeout(ctx, h.cfg.ExecutionTimeout)
	defer cancel()

	keyfunc := func(id peer.ID) string {
		return peerStrandKey(requestID, assignments[id].StrandID, id)
	}

	batchResults := gatherPeerMessages(
		waitctx,
		assignedWorkers,
		keyfunc,
		h.workOrderBatchResponses,
	)

	strandResults := make(map[string]response.NodeStrandResults)
	for peer, res := range batchResults {

		sr := response.NodeStrandResults{
			Peer:    peer,
			Results: res.Results,
		}

		assignment, ok := assignments[peer]
		// Should never happen.
		if !ok {
			return nil, fmt.Errorf("found a batch result for a peer without assignment (request: %v, peer: %v, reported strand id: %v)",
				requestID,
				peer.String(),
				res.StrandID)
		}

		strandResults[assignment.StrandID] = sr
	}

	return strandResults, nil
}

// generic helpers to get keys from a map. No locking or anything.
func mapKeys[K comparable, V any](m map[K]V) []K {

	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	return keys
}

func logAssignments(log *zerolog.Logger, assignments map[peer.ID]*request.WorkOrderBatch) {

	for peer, assignment := range assignments {
		log.Debug().
			Stringer("peer", peer).
			Int("count", len(assignment.Arguments)).
			Msg("work batch prepared for a peer")

		for i, args := range assignment.Arguments {
			log.Debug().
				Stringer("peer", peer).
				Int("i", i).
				Strs("arguments", args).
				Msg("work order variant")
		}
	}
}
