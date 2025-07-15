package batchstore

import (
	"context"
)

type Store interface {
	// TODO: Remove this from the interface.
	Init(context.Context) error

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
	CreateChunk(ctx context.Context, rec *ChunkRecord) error
	GetChunk(ctx context.Context, id string) (*ChunkRecord, error)
	UpdateChunk(ctx context.Context, rec *ChunkRecord) error
	UpdateChunkStatus(ctx context.Context, id string, status int32) error
	DeleteChunk(ctx context.Context, id string) error
}

type WorkItemStore interface {
	CreateWorkItem(ctx context.Context, rec *WorkItemRecord) error
	GetWorkItem(ctx context.Context, id string) (*WorkItemRecord, error)
	UpdateWorkItem(ctx context.Context, rec *WorkItemRecord) error
	UpdateWorkItemStatus(ctx context.Context, id string, status int32) error
	DeleteWorkItem(ctx context.Context, id string) error
}
