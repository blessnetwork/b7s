package mbs

import (
	"context"
	"fmt"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s *BatchStore) CreateWorkItem(ctx context.Context, rec *batchstore.WorkItemRecord) error {

	_, err := s.items.InsertOne(ctx, rec)
	if err != nil {
		return fmt.Errorf("could not insert work item: %w", err)
	}

	return nil
}

func (s *BatchStore) GetWorkItem(ctx context.Context, id string) (*batchstore.WorkItemRecord, error) {

	var item batchstore.WorkItemRecord
	err := s.items.FindOne(
		ctx,
		bson.D{{"id", id}},
	).Decode(&item)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve work item: %w", err)
	}

	return &item, nil
}

func (s *BatchStore) UpdateWorkItem(ctx context.Context, rec *batchstore.WorkItemRecord) error {

	_, err := s.items.UpdateOne(
		ctx,
		bson.D{{"id", rec.ID}},
		bson.D{{"$set", rec}},
	)
	if err != nil {
		return fmt.Errorf("could not update work item: %w", err)
	}

	return nil
}

func (s *BatchStore) UpdateWorkItemStatus(ctx context.Context, id string, status int32) error {

	_, err := s.items.UpdateOne(
		ctx,
		bson.D{{"id", id}},
		bson.D{{
			"$set",
			bson.D{{"status", status}}}},
	)
	if err != nil {
		return fmt.Errorf("could not update work item: %w", err)
	}

	return nil
}

func (s *BatchStore) DeleteWorkItem(ctx context.Context, id string) error {

	_, err := s.items.DeleteOne(ctx, bson.D{{"id", id}})
	if err != nil {
		return fmt.Errorf("could not delete work item: %w", err)
	}

	return nil
}
