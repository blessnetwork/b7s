package mbs

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	mongoNamespaceExistsCode = 48
)

type BatchStore struct {
	cfg Config
	cli *mongo.Client

	batches *mongo.Collection
	chunks  *mongo.Collection
	items   *mongo.Collection
}

func NewBatchStore(cli *mongo.Client, opts ...OptionFunc) (*BatchStore, error) {

	cfg := defaultConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	store := &BatchStore{
		cfg: cfg,
		cli: cli,
	}

	return store, nil
}

func (s *BatchStore) Init(ctx context.Context) error {

	// TODO: Move this stuff below to a separate method.

	if s.cfg.initCollections {
		err := s.createCollections(ctx)
		if err != nil {
			return fmt.Errorf("could not create collections: %w", err)
		}
	}

	db := s.cli.Database(s.cfg.dbname)
	s.batches = db.Collection(batchesCollection)
	s.chunks = db.Collection(chunksCollection)
	s.items = db.Collection(workItemCollection)

	return nil
}

func (s *BatchStore) createCollections(ctx context.Context) error {

	// TODO: Chunk or worker association?
	collections := map[string][]byte{
		batchesCollection:  batchCollectionSchema,
		chunksCollection:   chunkCollectionSchema,
		workItemCollection: workItemCollectionSchema,
	}

	for collection, schema := range collections {

		var compiled bson.M
		err := bson.UnmarshalExtJSON(schema, true, &compiled)
		if err != nil {
			return fmt.Errorf("invalid collection schema definition (collection: %v): %w", collection, err)
		}

		options := options.CreateCollection().SetValidator(compiled)

		// TODO: Use a specific database
		err = s.cli.Database("playground").CreateCollection(ctx, collection, options)

		// TODO: Honour config option.

		// NOTE: Because of how mongo (or the go driver) checks if the collection options are identical, we are a little less strict here.
		// Schema gets compiled to an unordered map and checked against the existing collection
		if err != nil && !isNamespaceExists(err) {
			return fmt.Errorf("could not create collection: %w", err)
		}
	}

	return nil
}

func isNamespaceExists(err error) bool {
	var se mongo.ServerError
	if !errors.As(err, &se) {
		return false
	}

	if se.HasErrorCode(mongoNamespaceExistsCode) {
		return true
	}

	return false
}
