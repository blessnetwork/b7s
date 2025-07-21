package batchstore

import (
	"context"
)

type Status int32

const (
	StatusCreated    = 0
	StatusInProgress = 1
	StatusFailed     = -1
	StatusDone       = 100
)

type Store interface {
	BatchStore
	ChunkStore
	WorkItemStore
}

// TODO: Use multiples instead of singles.

type BatchStore interface {
	CreateBatch(ctx context.Context, rec *ExecuteBatchRecord) error
	GetBatch(ctx context.Context, id string) (*ExecuteBatchRecord, error)
	UpdateBatch(ctx context.Context, rec *ExecuteBatchRecord) error
	UpdateBatchStatus(ctx context.Context, id string, status int32) error
	DeleteBatch(ctx context.Context, id string) error
}

type ChunkStore interface {
	CreateChunks(ctx context.Context, rec ...*ChunkRecord) error
	GetChunk(ctx context.Context, id string) (*ChunkRecord, error)
	UpdateChunk(ctx context.Context, rec *ChunkRecord) error
	UpdateChunkStatus(ctx context.Context, status int32, ids ...string) error
	DeleteChunks(ctx context.Context, ids ...string) error
	GetBatchChunks(ctx context.Context, batchID string) ([]*ChunkRecord, error)
}

type WorkItemStore interface {
	CreateWorkItems(ctx context.Context, rec ...*WorkItemRecord) error
	GetWorkItem(ctx context.Context, id string) (*WorkItemRecord, error)
	UpdateWorkItem(ctx context.Context, rec *WorkItemRecord) error
	UpdateWorkItemStatus(ctx context.Context, status int32, ids ...string) error
	DeleteWorkItems(ctx context.Context, ids ...string) error
	AssignWorkItems(ctx context.Context, chunkID string, ids ...string) error
	GetBatchWorkItems(ctx context.Context, batchID string) ([]*WorkItemRecord, error)
	GetChunkWorkItems(ctx context.Context, chunkID string) ([]*WorkItemRecord, error)
	GetBatchIncompleteWorkItems(ctx context.Context, batchID string) ([]*WorkItemRecord, error)

	// TODO: Perhaps create a single GetWorkItems function that accepts a query.
}
