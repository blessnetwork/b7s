package mocks

import (
	"context"
	"testing"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
)

var _ batchstore.Store = (*BatchStore)(nil)

type BatchStore struct {
	CreateBatchFunc       func(context.Context, *batchstore.ExecuteBatchRecord) error
	GetBatchFunc          func(context.Context, string) (*batchstore.ExecuteBatchRecord, error)
	UpdateBatchFunc       func(context.Context, *batchstore.ExecuteBatchRecord) error
	UpdateBatchStatusFunc func(context.Context, int32, string) error
	DeleteBatchFunc       func(context.Context, string) error
	FindBatchesFunc       func(context.Context, ...int32) ([]*batchstore.ExecuteBatchRecord, error)

	CreateChunksFunc      func(context.Context, ...*batchstore.ChunkRecord) error
	GetChunkFunc          func(context.Context, string) (*batchstore.ChunkRecord, error)
	GetBatchChunksFunc    func(context.Context, string) ([]*batchstore.ChunkRecord, error)
	UpdateChunkFunc       func(context.Context, *batchstore.ChunkRecord) error
	UpdateChunkStatusFunc func(context.Context, int32, ...string) error
	DeleteChunksFunc      func(context.Context, ...string) error

	CreateWorkItemsFunc      func(context.Context, ...*batchstore.WorkItemRecord) error
	GetWorkItemFunc          func(context.Context, string) (*batchstore.WorkItemRecord, error)
	UpdateWorkItemFunc       func(context.Context, *batchstore.WorkItemRecord) error
	UpdateWorkItemStatusFunc func(context.Context, int32, ...string) error
	DeleteWorkItemsFunc      func(context.Context, ...string) error
	AssignWorkItemsFunc      func(context.Context, string, ...string) error
	FindWorkItemsFunc        func(context.Context, string, string, ...int32) ([]*batchstore.WorkItemRecord, error)
}

// TODO: Add actual types to be returned, not nils

func BaselineMockStore(t *testing.T) *BatchStore {
	t.Helper()

	return &BatchStore{
		CreateBatchFunc: func(context.Context, *batchstore.ExecuteBatchRecord) error {
			return nil
		},
		GetBatchFunc: func(context.Context, string) (*batchstore.ExecuteBatchRecord, error) {
			return nil, nil
		},
		UpdateBatchFunc: func(context.Context, *batchstore.ExecuteBatchRecord) error {
			return nil
		},
		UpdateBatchStatusFunc: func(context.Context, int32, string) error {
			return nil
		},
		DeleteBatchFunc: func(context.Context, string) error {
			return nil
		},
		FindBatchesFunc: func(context.Context, ...int32) ([]*batchstore.ExecuteBatchRecord, error) {
			return nil, nil
		},

		CreateChunksFunc: func(context.Context, ...*batchstore.ChunkRecord) error {
			return nil
		},
		GetChunkFunc: func(context.Context, string) (*batchstore.ChunkRecord, error) {
			return nil, nil
		},
		GetBatchChunksFunc: func(context.Context, string) ([]*batchstore.ChunkRecord, error) {
			return nil, nil
		},
		UpdateChunkFunc: func(context.Context, *batchstore.ChunkRecord) error {
			return nil
		},
		UpdateChunkStatusFunc: func(context.Context, int32, ...string) error {
			return nil
		},
		DeleteChunksFunc: func(context.Context, ...string) error {
			return nil
		},

		CreateWorkItemsFunc: func(context.Context, ...*batchstore.WorkItemRecord) error {
			return nil
		},
		GetWorkItemFunc: func(context.Context, string) (*batchstore.WorkItemRecord, error) {
			return nil, nil
		},
		UpdateWorkItemFunc: func(context.Context, *batchstore.WorkItemRecord) error {
			return nil
		},
		UpdateWorkItemStatusFunc: func(context.Context, int32, ...string) error {
			return nil
		},
		DeleteWorkItemsFunc: func(context.Context, ...string) error {
			return nil
		},
		AssignWorkItemsFunc: func(context.Context, string, ...string) error {
			return nil
		},
		FindWorkItemsFunc: func(context.Context, string, string, ...int32) ([]*batchstore.WorkItemRecord, error) {
			return nil, nil
		},
	}
}

func (m *BatchStore) CreateBatch(ctx context.Context, rec *batchstore.ExecuteBatchRecord) error {
	return m.CreateBatchFunc(ctx, rec)
}

func (m *BatchStore) GetBatch(ctx context.Context, id string) (*batchstore.ExecuteBatchRecord, error) {
	return m.GetBatchFunc(ctx, id)
}

func (m *BatchStore) UpdateBatch(ctx context.Context, rec *batchstore.ExecuteBatchRecord) error {
	return m.UpdateBatchFunc(ctx, rec)
}

func (m *BatchStore) UpdateBatchStatus(ctx context.Context, status int32, id string) error {
	return m.UpdateBatchStatusFunc(ctx, status, id)
}

func (m *BatchStore) DeleteBatch(ctx context.Context, id string) error {
	return m.DeleteBatchFunc(ctx, id)
}

func (m *BatchStore) FindBatches(ctx context.Context, statuses ...int32) ([]*batchstore.ExecuteBatchRecord, error) {
	return m.FindBatchesFunc(ctx, statuses...)
}

func (m *BatchStore) CreateChunks(ctx context.Context, rec ...*batchstore.ChunkRecord) error {
	return m.CreateChunksFunc(ctx, rec...)
}

func (m *BatchStore) GetChunk(ctx context.Context, id string) (*batchstore.ChunkRecord, error) {
	return m.GetChunkFunc(ctx, id)
}

func (m *BatchStore) GetBatchChunks(ctx context.Context, batchID string) ([]*batchstore.ChunkRecord, error) {
	return m.GetBatchChunksFunc(ctx, batchID)
}

func (m *BatchStore) UpdateChunk(ctx context.Context, rec *batchstore.ChunkRecord) error {
	return m.UpdateChunkFunc(ctx, rec)
}

func (m *BatchStore) UpdateChunkStatus(ctx context.Context, status int32, ids ...string) error {
	return m.UpdateChunkStatusFunc(ctx, status, ids...)
}

func (m *BatchStore) DeleteChunks(ctx context.Context, ids ...string) error {
	return m.DeleteChunksFunc(ctx, ids...)
}

func (m *BatchStore) CreateWorkItems(ctx context.Context, rec ...*batchstore.WorkItemRecord) error {
	return m.CreateWorkItemsFunc(ctx, rec...)
}

func (m *BatchStore) GetWorkItem(ctx context.Context, id string) (*batchstore.WorkItemRecord, error) {
	return m.GetWorkItemFunc(ctx, id)
}

func (m *BatchStore) UpdateWorkItem(ctx context.Context, rec *batchstore.WorkItemRecord) error {
	return m.UpdateWorkItemFunc(ctx, rec)
}

func (m *BatchStore) UpdateWorkItemStatus(ctx context.Context, status int32, ids ...string) error {
	return m.UpdateWorkItemStatusFunc(ctx, status, ids...)
}

func (m *BatchStore) DeleteWorkItems(ctx context.Context, ids ...string) error {
	return m.DeleteWorkItemsFunc(ctx, ids...)
}

func (m *BatchStore) AssignWorkItems(ctx context.Context, chunkID string, ids ...string) error {
	return m.AssignWorkItemsFunc(ctx, chunkID, ids...)
}

func (m *BatchStore) FindWorkItems(ctx context.Context, batchID string, chunkID string, statuses ...int32) ([]*batchstore.WorkItemRecord, error) {
	return m.FindWorkItemsFunc(ctx, batchID, chunkID, statuses...)
}
