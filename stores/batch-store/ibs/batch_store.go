package ibs

import (
	"context"
	"errors"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
)

type BatchStore struct {
}

func NewBatchStore() (*BatchStore, error) {
	return &BatchStore{}, errors.New("TBD: Not implemented")
}

func (s *BatchStore) CreateBatch(ctx context.Context, rec *batchstore.ExecuteBatchRecord) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) GetBatch(ctx context.Context, id string) (*batchstore.ExecuteBatchRecord, error) {
	return nil, errors.New("TBD: not implemented")
}

func (s *BatchStore) UpdateBatch(ctx context.Context, rec *batchstore.ExecuteBatchRecord) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) UpdateBatchStatus(ctx context.Context, id string, status int32) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) DeleteBatch(ctx context.Context, id string) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) CreateChunk(ctx context.Context, rec *batchstore.ChunkRecord) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) GetChunk(ctx context.Context, id string) (*batchstore.ChunkRecord, error) {
	return nil, errors.New("TBD: not implemented")
}

func (s *BatchStore) UpdateChunk(ctx context.Context, rec *batchstore.ChunkRecord) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) UpdateChunkStatus(ctx context.Context, id string, status int32) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) DeleteChunk(ctx context.Context, id string) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) CreateWorkItem(ctx context.Context, rec *batchstore.WorkItemRecord) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) GetWorkItem(ctx context.Context, id string) (*batchstore.WorkItemRecord, error) {
	return nil, errors.New("TBD: not implemented")
}

func (s *BatchStore) UpdateWorkItem(ctx context.Context, rec *batchstore.WorkItemRecord) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) UpdateWorkItemStatus(ctx context.Context, id string, status int32) error {
	return errors.New("TBD: not implemented")
}

func (s *BatchStore) DeleteWorkItem(ctx context.Context, id string) error {
	return errors.New("TBD: not implemented")
}
