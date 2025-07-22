package head

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/models/response"
	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
)

// TODO: IDs are getting too big.
// TODO: RequestHash can be a byte sequence instead of a hash.

func workItemID(batchID string, itemID string) string {
	return batchID + "/" + itemID
}

func (h *HeadNode) saveBatch(id string, req request.ExecuteBatch) error {

	err := h.createBatch(id, req)
	if err != nil {
		return fmt.Errorf("could not persist batch: %w", err)
	}

	items := make([]*batchstore.WorkItemRecord, len(req.Arguments))
	for i, args := range req.Arguments {

		itemID := execute.ExecutionID(
			req.Template.FunctionID,
			req.Template.Method,
			args)

		items[i] = &batchstore.WorkItemRecord{
			ID:        workItemID(id, string(itemID)),
			RequestID: id,
			Arguments: args,
			Status:    0,
			Attempts:  0,
		}
	}

	err = h.cfg.BatchStore.CreateWorkItems(context.TODO(), items...)
	if err != nil {
		return fmt.Errorf("could not persist work items: %w", err)
	}

	return nil
}

func (h *HeadNode) createBatch(id string, req request.ExecuteBatch) error {

	rec := batchstore.ExecuteBatchRecord{
		ID:        id,
		CID:       req.Template.FunctionID,
		Method:    req.Template.Method,
		Config:    req.Template.Config,
		CreatedAt: time.Now().UTC(),
	}

	return h.cfg.BatchStore.CreateBatch(context.TODO(), &rec)
}

func (h *HeadNode) saveChunkInfo(id string, assignments map[peer.ID]*request.WorkOrderBatch) error {

	// Create chunk records.
	err := h.createChunks(id, assignments)
	if err != nil {
		return fmt.Errorf("could not save chunks: %w", err)
	}

	// Update work items to set their assignments (associate chunk ID).
	err = h.updateWorkOrderAssignments(id, assignments)
	if err != nil {
		return fmt.Errorf("could not update work items to assign chunk ID: %w", err)
	}

	return nil
}

func (h *HeadNode) createChunks(id string, assignments map[peer.ID]*request.WorkOrderBatch) error {

	ts := time.Now().UTC()
	chunks := make([]*batchstore.ChunkRecord, len(assignments))

	i := 0
	for _, chunk := range assignments {
		chunks[i] = &batchstore.ChunkRecord{
			ID:        chunk.ChunkID,
			RequestID: id,
			Status:    0,
			CreatedAt: ts,
		}

		i++
	}

	return h.cfg.BatchStore.CreateChunks(context.TODO(), chunks...)
}

func (h *HeadNode) updateWorkOrderAssignments(id string, assignments map[peer.ID]*request.WorkOrderBatch) error {

	for peer, chunk := range assignments {

		ids := make([]string, len(chunk.Arguments))
		for i, args := range chunk.Arguments {
			ids[i] = string(execute.ExecutionID(chunk.Template.FunctionID, chunk.Template.Method, args))
		}

		// NOTE: Potentially inefficient - one query per chunk.
		err := h.cfg.BatchStore.AssignWorkItems(context.TODO(), chunk.ChunkID, ids...)
		if err != nil {
			return fmt.Errorf("could not update work item assignment (chunk: %v, worker: %v): %w", chunk.ChunkID, peer.String(), err)
		}
	}

	return nil
}

func (h *HeadNode) markStartedChunks(id string, assignments map[peer.ID]*request.WorkOrderBatch, ignore []peer.ID) error {

	var multierr *multierror.Error
	for peer, chunk := range assignments {
		// Skip chunks to which deliveries failed.
		if slices.Contains(ignore, peer) {
			continue
		}

		ids := make([]string, len(chunk.Arguments))
		for i, args := range chunk.Arguments {
			ids[i] = string(execute.ExecutionID(chunk.Template.FunctionID, chunk.Template.Method, args))
		}

		err := h.cfg.BatchStore.UpdateWorkItemStatus(context.TODO(), batchstore.StatusInProgress, ids...)
		if err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}

	return multierr.ErrorOrNil()
}

func (h *HeadNode) markCompletedChunks(id string, chunkResults map[string]response.NodeChunkResults) error {

	// Group resulting work items by status so we can update them in batches.
	statuses := make(map[batchstore.Status][]string)
	for chunkID, res := range chunkResults {

		for itemID, itemResult := range res.Results {

			status := exitCodeToBatchStoreStatus(itemResult.Result.Result.ExitCode)

			h.Log().Debug().
				Str("batch", id).
				Str("chunk", chunkID).
				Str("item_id", string(itemID)).
				Int32("status", int32(status)).
				Int("exit_code", itemResult.Result.Result.ExitCode).
				Msg("processing chunk work item")

			_, ok := statuses[status]
			if !ok {
				statuses[status] = make([]string, 0, 10)
			}

			statuses[status] = append(statuses[status], string(itemID))
		}
	}

	var err *multierror.Error

	for status, ids := range statuses {

		log := h.Log().
			With().
			Str("batch", id).
			Int32("status", int32(status)).
			Strs("items", ids).
			Logger()

		log.Debug().Msg("updating work item status in batch store")

		err := h.cfg.BatchStore.UpdateWorkItemStatus(context.TODO(), int32(status), ids...)
		if err != nil {
			// Logging AND returning the message here but extra context is useful
			log.Error().Err(err).Msg("could not update work item status")

			err = multierror.Append(err, fmt.Errorf("could not update work item status: %w", err))
		}
	}

	return err.ErrorOrNil()
}

func exitCodeToBatchStoreStatus(e int) batchstore.Status {
	switch e {
	case 0:
		return batchstore.StatusDone
	default:
		return batchstore.StatusFailed
	}
}
