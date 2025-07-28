package head

import (
	"context"
	"errors"
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/models/response"
	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
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

	// Persist batch and work items.
	err := h.saveBatch(requestID, req)
	if err != nil {
		return fmt.Errorf("could not save batch request: %w", err)
	}

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

type batchResults map[string]response.NodeChunkResults

func (h *HeadNode) executeBatch(
	ctx context.Context,
	requestID string,
	req request.ExecuteBatch,
) (
	batchResults,
	error,
) {

	// TODO: Metrics and tracing

	log := h.Log().With().
		Str("request", requestID).
		Str("function", req.Template.FunctionID).
		Logger()

	log.Info().Msg("processing batch execution request")

	// Phase 1. - Issue roll call to nodes.

	rc := rollCallRequest(req.Template.FunctionID, requestID, 0, req.Template.Config.Attributes, true)

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

	// TODO: Perhaps we don't do this at all before chunks are actually sent?

	// XXX:
	// 1. create chunks in the DB
	// 2. update work items to contain chunk information to which they are assigned to.
	err = h.saveChunkInfo(requestID, assignments)
	if err != nil {
		return nil, fmt.Errorf("could not save chunks: %w", err)
	}

	// TODO: Rethink, useful but ugly.
	logAssignments(&log, assignments)

	var failedDeliveries []peer.ID
	err = h.sendBatch(ctx, assignments)
	if err != nil {

		var sendErr *batchSendError
		if errors.As(err, &sendErr) {
			// TODO: Handle partial failures by retrying part of the batch that failed.
			log.Warn().
				Strs("peers", bls.PeerIDsToStr(sendErr.Targets())).
				Msg("partial failure to send batch requst")

			failedDeliveries = sendErr.Targets()
		}

		return nil, fmt.Errorf("could not send work order batch: %w", err)
	}

	err = h.markStartedChunks(requestID, assignments, failedDeliveries)
	if err != nil {
		return nil, fmt.Errorf("could not mark chunks as in-progress: %w", err)
	}

	// TODO: Handle errors - reintroduce to the pool.

	// Wait for results.

	assignedWorkers := mapKeys(assignments)

	waitctx, cancel := context.WithTimeout(ctx, h.cfg.ExecutionTimeout)
	defer cancel()

	keyfunc := func(id peer.ID) string {
		return peerChunkKey(requestID, assignments[id].ChunkID, id)
	}

	batchResults := gatherPeerMessages(
		waitctx,
		assignedWorkers,
		keyfunc,
		h.workOrderBatchResponses,
	)

	chunkResults := make(map[string]response.NodeChunkResults)
	for peer, res := range batchResults {

		sr := response.NodeChunkResults{
			Peer:    peer,
			Results: res.Results,
		}

		assignment, ok := assignments[peer]
		// Should never happen.
		if !ok {
			return nil, fmt.Errorf("found a batch result for a peer without assignment (request: %v, peer: %v, reported chunk id: %v)",
				requestID,
				peer.String(),
				res.ChunkID)
		}

		chunkResults[assignment.ChunkID] = sr
	}

	err = h.markCompletedChunks(requestID, chunkResults)
	if err != nil {
		return nil, fmt.Errorf("could not mark chunks as complete: %w", err)
	}

	return chunkResults, nil
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

// Collect any work items for the batch that have not been executed yet and start their execution again.
func (h *HeadNode) continueBatchExecution(ctx context.Context, requestID string) error {

	log := h.Log().With().
		Str("request", requestID).
		Logger()

	log.Info().Msg("continuing batch execution")

	batch, err := h.cfg.BatchStore.GetBatch(ctx, requestID)
	if err != nil {
		return fmt.Errorf("could not retrieve batch: %w", err)
	}

	// NOTE: Perhaps we should not trust this flag.
	if batch.Status == batchstore.StatusDone {
		log.Info().Msg("batch reported as completed, stopping")
		return nil
	}

	// We want to restart execution of failed items, or those that were created but not started
	items, err := h.cfg.BatchStore.FindWorkItems(ctx, requestID, "", batchstore.StatusCreated, batchstore.StatusFailed)
	if err != nil {
		return fmt.Errorf("could not retrieve work items for batch (batch:%v): %w", requestID, err)
	}

	threshold := min(h.cfg.BatchWorkItemMaxAttempts, batch.MaxAttempts)
	pending, permaFailed := filterWorkItems(items, threshold)

	go func(ctx context.Context) {
		formatWorkRecordIDs := func(items []*batchstore.WorkItemRecord) []string {
			ids := make([]string, 0, len(items))
			for i, item := range items {
				ids[i] = workItemID(requestID, string(execute.ExecutionID(batch.CID, batch.Method, item.Arguments)))
			}

			return ids
		}
		err = h.cfg.BatchStore.UpdateWorkItemStatus(ctx, batchstore.StatusPermanentlyFailed, formatWorkRecordIDs(permaFailed)...)
		if err != nil {
			log.Error().Err(err).Msg("could not mark items as permanently failed")
		}
	}(ctx)

	// TODO: Do we need the return value?
	_, err = h.executeBatch(ctx, requestID, batchRecordToRequest(batch, pending))
	if err != nil {
		return fmt.Errorf("could not continue batch execution: %w", err)
	}

	return nil
}

// Split work item list into two categories:
// pending - created or failed ones
// perma failed - items that failed execution N times
func filterWorkItems(items []*batchstore.WorkItemRecord, threshold uint32) ([]*batchstore.WorkItemRecord, []*batchstore.WorkItemRecord) {

	var (
		pending     = make([]*batchstore.WorkItemRecord, 0, len(items))
		permaFailed []*batchstore.WorkItemRecord
	)

	for _, item := range items {

		switch item.Status {
		case batchstore.StatusCreated:
			pending = append(pending, item)

		case batchstore.StatusFailed:

			if item.Attempts >= uint32(threshold) {
				permaFailed = append(permaFailed, item)
				continue
			}

			pending = append(pending, item)
		}
	}

	return pending, permaFailed
}
