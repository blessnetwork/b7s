package mbs

import (
	"context"
	"errors"
	"fmt"
	"time"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *BatchStore) CreateChunks(ctx context.Context, chunks ...*batchstore.ChunkRecord) error {

	_, err := s.chunks.InsertMany(ctx, chunks)
	if err != nil {
		return fmt.Errorf("could not save chunk: %w", err)
	}

	return nil
}

func (s *BatchStore) GetChunk(ctx context.Context, id string) (*batchstore.ChunkRecord, error) {

	var rec batchstore.ChunkRecord
	err := s.chunks.FindOne(
		ctx,
		bson.M{"id": id},
	).Decode(&rec)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve chunk: %w", err)
	}

	return &rec, nil
}

func (s *BatchStore) GetBatchChunks(ctx context.Context, batchID string) ([]*batchstore.ChunkRecord, error) {
	return nil, errors.New("TBD: not implemented")
}

func (s *BatchStore) UpdateChunk(ctx context.Context, rec *batchstore.ChunkRecord) error {

	// modding input record
	rec.UpdatedAt = time.Now().UTC()

	_, err := s.chunks.UpdateOne(
		ctx,
		bson.M{"id": rec.ID},
		bson.M{"$set": rec},
	)
	if err != nil {
		return fmt.Errorf("could not update chunk: %w", err)
	}

	return nil
}

func (s *BatchStore) UpdateChunkStatus(ctx context.Context, status int32, ids ...string) error {

	_, err := s.chunks.UpdateMany(
		ctx,
		bson.M{"id": bson.M{"$in": ids}},
		bson.M{"$set": bson.M{
			"status":     status,
			"updated_at": time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("could not update chunk: %w", err)
	}

	return nil
}

func (s *BatchStore) DeleteChunks(ctx context.Context, ids ...string) error {

	_, err := s.chunks.DeleteMany(
		ctx,
		bson.M{"id": bson.M{"$in": ids}},
	)
	if err != nil {
		return fmt.Errorf("could not delete chunk: %w", err)
	}

	return nil
}

func (s *BatchStore) FindChunks(ctx context.Context, batchID string, statuses ...int32) ([]*batchstore.ChunkRecord, error) {

	if batchID == "" {
		return nil, errors.New("batch ID is required")
	}

	query := make(map[string]any)
	query["batch_id"] = batchID

	sn := len(statuses)
	if sn == 1 {
		query["status"] = statuses[0]
	} else if sn > 1 {
		query["status"] = map[string]any{
			"$in": statuses,
		}
	}

	cursor, err := s.chunks.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not lookup chunks: %w", err)
	}

	var chunks []*batchstore.ChunkRecord
	err = cursor.All(ctx, &chunks)
	if err != nil {
		return nil, fmt.Errorf("could not decode found chunks: %w", err)
	}

	return chunks, nil
}
