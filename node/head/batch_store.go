package head

import (
	"context"
	"fmt"
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

// itemID is a request hash which is derived from what we are tasked to execute. However this does not need
// to be universally unique. We can have multiple execution requests that are executing the same bless function
// with the same arguments. Hence we need the batchID as well.
func workItemID(batchID string, itemID string) string {
	return batchID + "/" + itemID
}

func (h *HeadNode) saveBatch(batchID string, req request.ExecuteBatch) error {

	batch, items := requestToBatchRecord(batchID, req)

	batch.Status = batchstore.StatusInProgress

	err := h.cfg.BatchStore.CreateBatch(context.TODO(), batch)
	if err != nil {
		return fmt.Errorf("could not persist batch: %w", err)
	}

	err = h.cfg.BatchStore.CreateWorkItems(context.TODO(), items...)
	if err != nil {
		return fmt.Errorf("could not persist work items: %w", err)
	}

	return nil
}

func (h *HeadNode) saveChunkInfo(batchID string, assignments map[peer.ID]*request.WorkOrderBatch) error {

	err := h.createChunks(batchID, assignments)
	if err != nil {
		return fmt.Errorf("could not save chunks: %w", err)
	}

	// Update work items to set their assignments (associate chunk ID).
	err = h.updateWorkOrderAssignments(batchID, assignments)
	if err != nil {
		return fmt.Errorf("could not update work items to assign chunk ID: %w", err)
	}

	return nil
}

func (h *HeadNode) createChunks(batchID string, assignments map[peer.ID]*request.WorkOrderBatch) error {

	ts := time.Now().UTC()
	chunks := make([]*batchstore.ChunkRecord, len(assignments))

	i := 0
	for _, chunk := range assignments {
		chunks[i] = &batchstore.ChunkRecord{
			ID:        chunk.ChunkID,
			BatchID:   batchID,
			Status:    batchstore.StatusCreated,
			CreatedAt: ts,
		}

		i++
	}

	return h.cfg.BatchStore.CreateChunks(context.TODO(), chunks...)
}

func (h *HeadNode) updateWorkOrderAssignments(batchID string, assignments map[peer.ID]*request.WorkOrderBatch) error {

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

func (h *HeadNode) markStartedChunks(batchID string, assignments map[peer.ID]*request.WorkOrderBatch, ignore []peer.ID) error {

	// Faster lookup of chunks to not update.
	im := make(map[peer.ID]struct{})
	for _, peer := range ignore {
		im[peer] = struct{}{}
	}

	// Get list of started chunks so we can update them.
	started := make([]string, 0, len(assignments))
	for peer, chunk := range assignments {
		_, undelivered := im[peer]
		if undelivered {
			continue
		}

		started = append(started, chunk.ChunkID)
	}

	var multierr *multierror.Error

	// Update chunks all at once.
	err := h.cfg.BatchStore.UpdateChunkStatus(context.TODO(), batchstore.StatusInProgress, started...)
	if err != nil {
		multierr = multierror.Append(multierr, err)
	}

	// Update work items in larger batches.
	for peer, chunk := range assignments {

		_, undelivered := im[peer]
		if undelivered {
			continue
		}

		ids := make([]string, len(chunk.Arguments))
		for i, args := range chunk.Arguments {
			ids[i] = workItemID(batchID, string(execute.ExecutionID(chunk.Template.FunctionID, chunk.Template.Method, args)))
		}

		err := h.cfg.BatchStore.UpdateWorkItemStatus(context.TODO(), batchstore.StatusInProgress, ids...)
		if err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}

	return multierr.ErrorOrNil()
}

func (h *HeadNode) markCompletedChunks(batchID string, sizes map[string]int, chunkResults map[string]response.NodeChunkResults) error {

	// TODO: Mark chunk as complete if we have responses for all work items in that chunk.

	// Group resulting work items by status so we can update them in batches.
	completed := make([]string, 0, len(chunkResults))
	statuses := make(map[batchstore.Status][]string)
	for chunkID, res := range chunkResults {

		for itemID, itemResult := range res.Results {

			status := exitCodeToBatchStoreStatus(itemResult.Result.Result.ExitCode)

			h.Log().Debug().
				Str("batch", batchID).
				Str("chunk", chunkID).
				Str("item_id", string(itemID)).
				Int32("status", int32(status)).
				Int("exit_code", itemResult.Result.Result.ExitCode).
				Msg("processing chunk work item")

			_, ok := statuses[status]
			if !ok {
				statuses[status] = make([]string, 0, 10)
			}

			statuses[status] = append(statuses[status], workItemID(batchID, string(itemID)))
		}

		// If we have all of the results - mark the chunk as done.
		if len(res.Results) == sizes[chunkID] {
			completed = append(completed, chunkID)
		}
	}

	var merr *multierror.Error

	for status, ids := range statuses {

		log := h.Log().
			With().
			Str("batch", batchID).
			Int32("status", int32(status)).
			Strs("items", ids).
			Logger()

		log.Debug().Msg("updating work item status in batch store")

		err := h.cfg.BatchStore.UpdateWorkItemStatus(context.TODO(), int32(status), ids...)
		if err != nil {
			// Logging AND returning the message here but extra context is useful
			log.Error().Err(err).Msg("could not update work item status")

			merr = multierror.Append(merr, fmt.Errorf("could not update work item status: %w", err))
		}
	}

	err := h.cfg.BatchStore.UpdateChunkStatus(context.TODO(), batchstore.StatusDone, completed...)
	if err != nil {
		h.Log().Error().
			Err(err).
			Strs("ids", completed).
			Msg("could not update chunk statutes")

		err = multierror.Append(merr, fmt.Errorf("could not update chunk status: %w", err))
	}

	return merr.ErrorOrNil()
}

func exitCodeToBatchStoreStatus(e int) batchstore.Status {
	switch e {
	case 0:
		return batchstore.StatusDone
	default:
		return batchstore.StatusFailed
	}
}
