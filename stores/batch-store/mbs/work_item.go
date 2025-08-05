package mbs

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
)

func (s *BatchStore) CreateWorkItems(ctx context.Context, rec ...*batchstore.WorkItemRecord) error {

	_, err := s.items.InsertMany(ctx, rec)
	if err != nil {
		return fmt.Errorf("could not insert work item: %w", err)
	}

	return nil
}

func (s *BatchStore) GetWorkItem(ctx context.Context, id string) (*batchstore.WorkItemRecord, error) {

	var item batchstore.WorkItemRecord
	err := s.items.FindOne(
		ctx,
		bson.M{"id": id},
	).Decode(&item)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve work item: %w", err)
	}

	return &item, nil
}

func (s *BatchStore) UpdateWorkItem(ctx context.Context, rec *batchstore.WorkItemRecord) error {

	// modding input record
	rec.UpdatedAt = time.Now().UTC()

	_, err := s.items.UpdateOne(
		ctx,
		bson.M{"id": rec.ID},
		bson.M{"$set": rec},
	)
	if err != nil {
		return fmt.Errorf("could not update work item: %w", err)
	}

	return nil
}

func (s *BatchStore) UpdateWorkItemStatus(ctx context.Context, status int32, ids ...string) error {

	query := bson.M{"$set": bson.M{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}}

	if status == batchstore.StatusFailed {
		query["$inc"] = bson.M{
			"attempts": 1,
		}
	}

	_, err := s.items.UpdateMany(
		ctx,
		bson.M{"id": bson.M{"$in": ids}},
		query,
	)
	if err != nil {
		return fmt.Errorf("could not update work item: %w", err)
	}

	return nil
}

func (s *BatchStore) FindWorkItems(ctx context.Context, batchID string, chunkID string, statuses ...int32) ([]*batchstore.WorkItemRecord, error) {

	query := make(map[string]any)

	if batchID != "" {
		query["batch_id"] = batchID
	}

	if chunkID != "" {
		query["chunk_id"] = chunkID
	}

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

	cursor, err := s.items.Find(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("could not lookup items: %w", err)
	}

	var items []*batchstore.WorkItemRecord
	err = cursor.All(ctx, &items)
	if err != nil {
		return nil, fmt.Errorf("could not decode found work items: %w", err)
	}

	return items, nil
}

func (s *BatchStore) DeleteWorkItems(ctx context.Context, ids ...string) error {

	_, err := s.items.DeleteMany(
		ctx,
		bson.M{"id": bson.M{"$in": ids}},
	)
	if err != nil {
		return fmt.Errorf("could not delete work item: %w", err)
	}

	return nil
}

// TODO: If there's too many IDs query can get troubling and we might need to consider chunking up the input list.
func (s *BatchStore) AssignWorkItems(ctx context.Context, chunkID string, ids ...string) error {

	_, err := s.items.UpdateMany(
		ctx,
		bson.M{"id": bson.M{"$in": ids}},
		bson.M{"$set": bson.M{
			"chunk_id":   chunkID,
			"status":     batchstore.StatusInProgress,
			"updated_at": time.Now().UTC(),
		}},
	)
	if err != nil {
		return fmt.Errorf("could not assign work item: %w", err)
	}

	return nil
}
