package ibs

import (
	"context"
	"errors"
	"sync"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
)

var _ batchstore.Store = (*BatchStore)(nil)

type BatchStore struct {
	*sync.RWMutex

	batches map[string]*batchstore.ExecuteBatchRecord
	chunks  map[string]*batchstore.ChunkRecord
	items   map[string]*batchstore.WorkItemRecord
}

func NewBatchStore() *BatchStore {
	bs := &BatchStore{
		RWMutex: &sync.RWMutex{},
		batches: make(map[string]*batchstore.ExecuteBatchRecord),
		chunks:  make(map[string]*batchstore.ChunkRecord),
		items:   make(map[string]*batchstore.WorkItemRecord),
	}

	return bs
}

func (s *BatchStore) CreateBatch(ctx context.Context, rec *batchstore.ExecuteBatchRecord) error {
	s.Lock()
	defer s.Unlock()

	s.batches[rec.ID] = rec
	return nil
}

func (s *BatchStore) GetBatch(ctx context.Context, id string) (*batchstore.ExecuteBatchRecord, error) {
	s.RLock()
	defer s.RUnlock()

	rec, ok := s.batches[id]
	if !ok {
		return nil, errors.New("batch not found")
	}

	return rec, nil
}

func (s *BatchStore) UpdateBatch(ctx context.Context, rec *batchstore.ExecuteBatchRecord) error {
	s.Lock()
	defer s.Unlock()

	s.batches[rec.ID] = rec
	return nil
}

func (s *BatchStore) UpdateBatchStatus(ctx context.Context, status int32, id string) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.batches[id]
	if !ok {
		return errors.New("batch not found")
	}

	s.batches[id].Status = status
	return nil
}

func (s *BatchStore) DeleteBatch(ctx context.Context, id string) error {
	s.Lock()
	defer s.Unlock()

	delete(s.batches, id)

	return nil
}

func (s *BatchStore) FindBatches(ctx context.Context, statuses ...int32) ([]*batchstore.ExecuteBatchRecord, error) {
	s.RLock()
	defer s.RUnlock()

	lookup := make(map[int32]struct{})
	for _, s := range statuses {
		lookup[s] = struct{}{}
	}

	var batches []*batchstore.ExecuteBatchRecord
	for _, batch := range s.batches {

		if len(lookup) > 0 {
			_, ok := lookup[batch.Status]
			if !ok {
				continue
			}
		}

		batches = append(batches, batch)
	}

	return batches, nil
}

func (s *BatchStore) CreateChunks(ctx context.Context, chunks ...*batchstore.ChunkRecord) error {
	s.Lock()
	defer s.Unlock()

	for _, chunk := range chunks {
		s.chunks[chunk.ID] = chunk
	}

	return nil
}

func (s *BatchStore) GetChunk(ctx context.Context, id string) (*batchstore.ChunkRecord, error) {
	s.RLock()
	defer s.RUnlock()

	rec, ok := s.chunks[id]
	if !ok {
		return nil, errors.New("chunk not found")
	}

	return rec, nil
}

func (s *BatchStore) GetBatchChunks(ctx context.Context, batchID string) ([]*batchstore.ChunkRecord, error) {
	s.RLock()
	defer s.RUnlock()

	var results []*batchstore.ChunkRecord
	for _, chunk := range s.chunks {
		if chunk.BatchID == batchID {
			results = append(results, chunk)
		}
	}

	return results, nil
}

func (s *BatchStore) UpdateChunk(ctx context.Context, rec *batchstore.ChunkRecord) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.chunks[rec.ID]
	if !ok {
		return errors.New("chunk not found")
	}

	s.chunks[rec.ID] = rec

	return nil
}

func (s *BatchStore) UpdateChunkStatus(ctx context.Context, status int32, ids ...string) error {
	s.Lock()
	defer s.Unlock()

	for _, id := range ids {

		_, ok := s.chunks[id]
		if !ok {
			return errors.New("chunk not found")
		}

		s.chunks[id].Status = status
	}

	return nil
}

func (s *BatchStore) DeleteChunks(ctx context.Context, ids ...string) error {
	s.Lock()
	defer s.Unlock()

	for _, id := range ids {
		delete(s.chunks, id)
	}

	return nil
}

func (s *BatchStore) CreateWorkItems(ctx context.Context, items ...*batchstore.WorkItemRecord) error {
	s.Lock()
	defer s.Unlock()

	for _, rec := range items {
		s.items[rec.ID] = rec
	}

	return nil
}

func (s *BatchStore) GetWorkItem(ctx context.Context, id string) (*batchstore.WorkItemRecord, error) {
	s.RLock()
	defer s.RUnlock()

	rec, ok := s.items[id]
	if !ok {
		return nil, errors.New("work item not found")
	}

	return rec, nil
}

func (s *BatchStore) UpdateWorkItem(ctx context.Context, rec *batchstore.WorkItemRecord) error {
	s.Lock()
	defer s.Unlock()

	_, ok := s.items[rec.ID]
	if !ok {
		return errors.New("work item not found")
	}

	s.items[rec.ID] = rec

	return nil
}

func (s *BatchStore) UpdateWorkItemStatus(ctx context.Context, status int32, ids ...string) error {
	s.Lock()
	defer s.Unlock()

	for _, id := range ids {

		_, ok := s.items[id]
		if !ok {
			return errors.New("work item not found")
		}

		s.items[id].Status = status
	}

	return nil
}

func (s *BatchStore) DeleteWorkItems(ctx context.Context, ids ...string) error {
	s.Lock()
	defer s.Unlock()

	for _, id := range ids {
		delete(s.items, id)
	}

	return nil
}

func (s *BatchStore) AssignWorkItems(ctx context.Context, chunkID string, ids ...string) error {
	s.Lock()
	defer s.Unlock()

	for _, id := range ids {
		_, ok := s.items[id]
		if !ok {
			return errors.New("item not found")
		}

		s.items[id].ChunkID = chunkID
	}

	return nil
}

func (s *BatchStore) FindWorkItems(ctx context.Context, batchID string, chunkID string, statuses ...int32) ([]*batchstore.WorkItemRecord, error) {
	s.RLock()
	defer s.RUnlock()

	lookup := make(map[int32]struct{})
	for _, s := range statuses {
		lookup[s] = struct{}{}
	}

	var results []*batchstore.WorkItemRecord
	for _, item := range s.items {

		if batchID != "" && item.BatchID != batchID {
			continue
		}

		if chunkID != "" && item.ChunkID != chunkID {
			continue
		}

		if len(lookup) > 0 {
			_, ok := lookup[item.Status]
			if !ok {
				continue
			}
		}

		results = append(results, item)
	}

	return results, nil
}
