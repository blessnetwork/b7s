package mbs

import (
	"context"
	"fmt"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *BatchStore) CreateBatch(ctx context.Context, rec *batchstore.ExecuteBatchRecord) error {

	_, err := s.batches.InsertOne(ctx, rec)
	if err != nil {
		return fmt.Errorf("could not save batch: %w", err)
	}

	return nil
}

func (s *BatchStore) GetBatch(ctx context.Context, id string) (*batchstore.ExecuteBatchRecord, error) {

	var rec batchstore.ExecuteBatchRecord
	err := s.batches.FindOne(
		ctx,
		bson.D{{"id", id}},
	).Decode(&rec)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve batch: %w", err)
	}

	return &rec, nil
}

func (s *BatchStore) UpdateBatch(ctx context.Context, rec *batchstore.ExecuteBatchRecord) error {

	_, err := s.batches.UpdateOne(
		ctx,
		bson.D{{"id", rec.ID}},
		bson.D{{"$set", rec}},
	)
	if err != nil {
		return fmt.Errorf("could not update batch: %w", err)
	}

	return nil
}

func (s *BatchStore) UpdateBatchStatus(ctx context.Context, id string, status int32) error {

	_, err := s.batches.UpdateOne(
		ctx,
		bson.D{{"id", id}},
		bson.D{{"$set",
			bson.D{{"status", status}}},
		},
	)
	if err != nil {
		return fmt.Errorf("could not update batch status: %w", err)
	}

	return nil
}

func (s *BatchStore) DeleteBatch(ctx context.Context, id string) error {

	_, err := s.batches.DeleteOne(
		ctx,
		bson.D{{"id", id}},
	)
	if err != nil {
		return fmt.Errorf("could not delete batch: %w", err)
	}

	return nil
}
