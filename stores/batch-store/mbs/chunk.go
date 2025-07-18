package mbs

import (
	"context"
	"fmt"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *BatchStore) CreateChunks(ctx context.Context, chunk ...*batchstore.ChunkRecord) error {

	_, err := s.chunks.InsertMany(ctx, chunk)
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

func (s *BatchStore) UpdateChunk(ctx context.Context, rec *batchstore.ChunkRecord) error {

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
		bson.M{"$set": bson.M{"status": status}},
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
