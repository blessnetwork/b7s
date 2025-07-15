package mbs

import (
	"context"
	"fmt"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *BatchStore) CreateChunk(ctx context.Context, chunk *batchstore.ChunkRecord) error {

	_, err := s.chunks.InsertOne(ctx, chunk)
	if err != nil {
		return fmt.Errorf("could not save chunk: %w", err)
	}

	return nil
}

func (s *BatchStore) GetChunk(ctx context.Context, id string) (*batchstore.ChunkRecord, error) {

	var rec batchstore.ChunkRecord
	err := s.chunks.FindOne(
		ctx,
		bson.D{{"id", id}},
	).Decode(&rec)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve chunk: %w", err)
	}

	return &rec, nil
}

func (s *BatchStore) UpdateChunk(ctx context.Context, rec *batchstore.ChunkRecord) error {

	_, err := s.chunks.UpdateOne(
		ctx,
		bson.D{{"id", rec.ID}},
		bson.D{{"$set", rec}},
	)
	if err != nil {
		return fmt.Errorf("could not update chunk: %w", err)
	}

	return nil
}

func (s *BatchStore) UpdateChunkStatus(ctx context.Context, id string, status int32) error {

	_, err := s.chunks.UpdateOne(
		ctx,
		bson.D{{"id", id}},
		bson.D{{"$set",
			bson.D{{"status", status}}}},
	)
	if err != nil {
		return fmt.Errorf("could not update chunk: %w", err)
	}

	return nil
}

func (s *BatchStore) DeleteChunk(ctx context.Context, id string) error {

	_, err := s.chunks.DeleteOne(
		ctx,
		bson.D{{"id", id}})
	if err != nil {
		return fmt.Errorf("could not delete chunk: %w", err)
	}

	return nil
}
