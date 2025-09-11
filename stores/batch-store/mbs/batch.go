package mbs

import (
	"context"
	"fmt"
	"time"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// TODO: Handle timestamps correctly.

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
		bson.M{"id": id},
	).Decode(&rec)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve batch: %w", err)
	}

	return &rec, nil
}

func (s *BatchStore) UpdateBatch(ctx context.Context, rec *batchstore.ExecuteBatchRecord) error {

	// modding input record
	rec.UpdatedAt = time.Now().UTC()

	_, err := s.batches.UpdateOne(
		ctx,
		bson.M{"id": rec.ID},
		bson.M{"$set": rec},
	)
	if err != nil {
		return fmt.Errorf("could not update batch: %w", err)
	}

	return nil
}

func (s *BatchStore) UpdateBatchStatus(ctx context.Context, status int32, id string) error {

	_, err := s.batches.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{"$set": bson.M{
			"status":     status,
			"updated_at": time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("could not update batch status: %w", err)
	}

	return nil
}

func (s *BatchStore) DeleteBatch(ctx context.Context, id string) error {

	_, err := s.batches.DeleteOne(
		ctx,
		bson.M{"id": id},
	)
	if err != nil {
		return fmt.Errorf("could not delete batch: %w", err)
	}

	return nil
}

func (s *BatchStore) FindBatches(ctx context.Context, statuses ...int32) ([]*batchstore.ExecuteBatchRecord, error) {

	query := make(map[string]any)

	sn := len(statuses)
	if sn == 1 {
		// Exact match for status
		query["status"] = statuses[0]
	} else if sn > 1 {
		// We have a list of statuses.
		query["status"] = map[string]any{
			"$in": statuses,
		}
	}

	cursor, err := s.batches.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not lookup batches: %w", err)
	}

	var batches []*batchstore.ExecuteBatchRecord
	err = cursor.All(ctx, &batches)
	if err != nil {
		return nil, fmt.Errorf("could not decode found batches: %w", err)
	}

	return batches, nil
}
