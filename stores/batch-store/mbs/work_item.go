package mbs

import (
	"context"
	"fmt"

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

	_, err := s.items.UpdateMany(
		ctx,
		bson.M{"id": bson.M{"$in": ids}},
		bson.M{"$set": bson.M{"status": status}},
	)
	if err != nil {
		return fmt.Errorf("could not update work item: %w", err)
	}

	return nil
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
