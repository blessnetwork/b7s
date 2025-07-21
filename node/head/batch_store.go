package head

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"github.com/libp2p/go-libp2p/core/peer"
)

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

func (h *HeadNode) updateChunkStatus(id string, assignments map[peer.ID]*request.WorkOrderBatch, status int32) error {
	return errors.New("TBD: not implemented")
}
