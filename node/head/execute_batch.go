package head

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/libp2p/go-libp2p/core/peer"

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

	log.Info().Msg("received batch execution request")

	// Persist batch and work items.
	err := h.saveBatch(requestID, req)
	if err != nil {
		return fmt.Errorf("could not save batch request: %w", err)
	}

	// TODO: Reset "in progress" work items: Head node could send a chunk to a worker,
	// and the worker might crash or something. The head node will then consider that chunk as "in progress"
	// as it was delivered to the worker, while the worker will (after restart) lose any info about the chunk
	// it was sent. When looking into resuming a batch - we must consider these work items -
	// - at certain point they should be reset and no longer be considered "in progress".

	// TODO: When a work order batch response is received out of band, status should be updated too.
	err = h.startBatchExecution(ctx, requestID, req)
	if err != nil {
		return fmt.Errorf("could not execute batch request: %w", err)
	}

	log.Info().Msg("started batch execution")

	res := req.Response(codes.OK, requestID)

	err = h.Send(ctx, from, res)
	if err != nil {
		return fmt.Errorf("could not send batch response: %w", err)
	}

	return nil
}

func (h *HeadNode) startBatchExecution(
	ctx context.Context,
	requestID string,
	req request.ExecuteBatch,
) error {

	// TODO: Metrics

	log := h.Log().With().
		Str("request", requestID).
		Str("function", req.Template.FunctionID).
		Logger()

	log.Info().Msg("processing batch execution request")

	// Phase 1. - Issue roll call to nodes.
	rc := req.RollCall(requestID)

	rctx, cancel := context.WithTimeout(ctx, h.cfg.ExecutionTimeout)
	defer cancel()

	// node count is -1 - we want all the nodes that want to work.
	peers, err := h.executeRollCall(rctx, rc, req.Topic, req.Template.Config.NodeCount)
	if err != nil {
		return fmt.Errorf("could not execute roll call: %w", err)
	}

	log.Debug().
		Strs("peers", bls.PeerIDsToStr(peers)).
		Msg("peers reported for work")

	assignments := partitionWorkBatch(peers, requestID, req)

	// XXX:
	// 1. create chunks in the DB
	// 2. update work items to contain chunk information to which they are assigned to.
	err = h.saveChunkInfo(requestID, assignments)
	if err != nil {
		return fmt.Errorf("could not save chunks: %w", err)
	}

	// Useful but ugly, won't use it in normal operation unless it proves to be required.
	// logAssignments(&log, assignments)

	var failedDeliveries []peer.ID
	err = h.sendBatch(ctx, assignments)
	if err != nil {

		var sendErr *batchSendError
		if !errors.As(err, &sendErr) {
			return fmt.Errorf("could not send work order batch: %w", err)
		}

		log.Warn().
			Strs("peers", bls.PeerIDsToStr(sendErr.Targets())).
			Msg("partial failure to send batch requst")

		failedDeliveries = sendErr.Targets()
	}

	err = h.markStartedChunks(requestID, assignments, failedDeliveries)
	if err != nil {
		return fmt.Errorf("could not mark chunks as in-progress: %w", err)
	}

	return nil
}

// // generic helpers to get keys from a map. No locking or anything.
// func mapKeys[K comparable, V any](m map[K]V) []K {

// 	keys := make([]K, 0, len(m))
// 	for key := range m {
// 		keys = append(keys, key)
// 	}

// 	return keys
// }
//
// func logAssignments(log *zerolog.Logger, assignments map[peer.ID]*request.WorkOrderBatch) {
//
// 		log.Debug().
// 			Stringer("peer", peer).
// 			Int("count", len(assignment.Arguments)).
// 			Msg("work batch prepared for a peer")
//
// 		for i, args := range assignment.Arguments {
// 			log.Debug().
// 				Stringer("peer", peer).
// 				Int("i", i).
// 				Strs("arguments", args).
// 				Msg("work order variant")
// 		}
// 	}
// }

// Collect any work items for the batch that have not been executed yet and start their execution again.
func (h *HeadNode) continueBatchExecution(ctx context.Context, batch *batchstore.ExecuteBatchRecord) error {

	requestID := batch.ID

	log := h.Log().With().Str("batch", batch.ID).Logger()

	log.Info().Msg("continuing batch execution")

	if batch.Status == batchstore.StatusDone {
		log.Info().Msg("batch reported as completed, stopping")
		return nil
	}

	// We want to restart execution of failed items, or those that were created but not started
	items, err := h.cfg.BatchStore.FindWorkItems(ctx, requestID, "", batchstore.StatusCreated, batchstore.StatusFailed)
	if err != nil {
		return fmt.Errorf("could not retrieve work items for batch (batch:%v): %w", requestID, err)
	}

	threshold := min(h.cfg.WorkItemMaxAttempts, batch.MaxAttempts)
	pending, permaFailed := filterWorkItems(items, threshold)

	if len(permaFailed) > 0 {
		go func(ctx context.Context) {

			h.Log().Info().Str("batch", requestID).Int("count", len(permaFailed)).
				Msg("marking work items as permanently failed")

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
	}

	if len(pending) == 0 {

		h.Log().Info().Str("batch", requestID).
			Msg("no pending work items - marking batch as done")

		return h.cfg.BatchStore.UpdateBatchStatus(ctx, batchstore.StatusDone, requestID)
	}

	h.Log().Info().Str("batch", requestID).Int("pending", len(pending)).
		Msg("requeuing batch work items")

	// TODO: We should no longer use the original number of nodes - we might only be processing 2% of work items, no reason to request the original N number of workers.
	err = h.startBatchExecution(ctx, requestID, batchRecordToRequest(batch, pending))
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

func (h *HeadNode) processWorkOrderBatchResponse(ctx context.Context, from peer.ID, res response.WorkOrderBatch) error {

	log := h.Log().With().
		Stringer("from", from).
		Str("batch", res.RequestID).
		Str("chunk", res.ChunkID).
		Logger()

	log.Debug().Msg("received work order batch response")

	// Perhaps on batch resume, node should first check the batch response cache and update the status for those work items.

	chunk, err := h.cfg.BatchStore.GetChunk(ctx, res.ChunkID)
	if err != nil {
		return fmt.Errorf("no matching chunk found (batch: %v, chunk: %v, peer: %v)", res.RequestID, res.ChunkID, from.String())
	}

	if chunk.Worker != from.String() {
		return fmt.Errorf("unexpected worker returned result (chunk: %v, expected: %v, got: %v)", res.RequestID, chunk.Worker, from.String())
	}

	// We'll convert this into a map as it's a more usable format for what we need.
	statuses := make(map[string]batchstore.WorkItemStatus)
	for itemID, itemResult := range res.Results {

		status := exitCodeToBatchStoreStatus(itemResult.Result.Result.ExitCode)

		log.Debug().
			Str("item_id", string(itemID)).
			Int32("status", int32(status)).
			Int("exit_code", itemResult.Result.Result.ExitCode).
			Msg("processing chunk work item")

		statuses[workItemID(chunk.BatchID, string(itemID))] = batchstore.WorkItemStatus{
			Status: status,
			Output: itemResult.Result.Result.Stdout,
		}
	}

	// Now that we have a map - we can use it for lookup to make sure all items we have are eligible to be updated.
	// For example - all work items this node returned actually do belong to this chunk.
	items, err := h.cfg.BatchStore.FindWorkItems(ctx, "", res.ChunkID, batchstore.StatusInProgress)
	if err != nil {
		return fmt.Errorf("could not retrieve chunk work items (chunk: %v): %w", res.ChunkID, err)
	}

	for _, item := range items {
		if res.ChunkID != item.ChunkID {
			return fmt.Errorf("item received from worker belongs to a different chunk (item: %v received_chunk: %v, actual: %v)",
				item.ID, res.ChunkID, item.ChunkID)
		}
	}

	h.Log().Info().
		Int("count", len(statuses)).
		Msg("updating work item status in batch store")

	var merr *multierror.Error

	err = h.cfg.BatchStore.UpdateWorkItemsOutput(ctx, statuses)
	if err != nil {
		// Logging AND returning the message here but extra context is useful
		log.Error().Err(err).Msg("could not update work item status")

		merr = multierror.Append(merr, fmt.Errorf("could not update work item status: %w", err))
	}

	if len(items) == len(statuses) {
		err = h.cfg.BatchStore.UpdateChunkStatus(ctx, batchstore.StatusDone, res.ChunkID)
		if err != nil {
			// Logging AND returning the message here but extra context is useful
			log.Error().Err(err).Msg("could not update chunk status")

			merr = multierror.Append(merr, fmt.Errorf("could not update chunk status: %w", err))
		}
	}

	return merr.ErrorOrNil()
}
