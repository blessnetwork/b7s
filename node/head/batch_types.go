package head

import (
	"time"

	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
)

// Convert a batchstore record format to request format.
func batchRecordToRequest(batch *batchstore.ExecuteBatchRecord, items []*batchstore.WorkItemRecord) request.ExecuteBatch {

	args := make([][]string, len(items))
	for i, item := range items {
		args[i] = item.Arguments
	}

	req := request.ExecuteBatch{
		Topic: "", // TODO: Add topic support
		Template: request.ExecutionRequestTemplate{
			FunctionID: batch.CID,
			Method:     batch.Method,
			Config:     execute.Config(batch.Config),
		},
		Arguments:   args,
		MaxAttempts: batch.MaxAttempts,
	}

	return req
}

// Convert request format to batchstore record format.
func requestToBatchRecord(id string, req request.ExecuteBatch) (*batchstore.ExecuteBatchRecord, []*batchstore.WorkItemRecord) {

	batch := &batchstore.ExecuteBatchRecord{
		ID:        id,
		CID:       req.Template.FunctionID,
		Method:    req.Template.Method,
		Config:    req.Template.Config,
		Status:    batchstore.StatusCreated,
		CreatedAt: time.Now().UTC(),
	}

	items := make([]*batchstore.WorkItemRecord, len(req.Arguments))
	for i, args := range req.Arguments {

		itemID := execute.ExecutionID(
			req.Template.FunctionID,
			req.Template.Method,
			args)

		items[i] = &batchstore.WorkItemRecord{
			ID:        workItemID(id, string(itemID)),
			BatchID:   id,
			Arguments: args,
			Status:    batchstore.StatusCreated,
			Attempts:  0,
		}
	}

	return batch, items
}
