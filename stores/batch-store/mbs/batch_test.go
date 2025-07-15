//go:build integration
// +build integration

package mbs_test

import (
	"fmt"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"

	b7smongo "github.com/blessnetwork/b7s/mongo"
	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"github.com/blessnetwork/b7s/stores/batch-store/mbs"
)

const (
	MongoDBConnectionEnv = "B7S_MONGO_DB_ADDRESS"
)

func TestBatchStore(t *testing.T) {

	var (
		client = getDBClient(t)
		ctx    = t.Context()
	)

	store, err := mbs.NewBatchStore(client)
	require.NoError(t, err)

	err = store.Init(ctx)
	require.NoError(t, err)

	batch := batchstore.ExecuteBatchRecord{
		ID:     uuid.New().String(),
		CID:    "test-cid",
		Method: "method.wasm",
		Status: 0,
	}

	t.Run("create", func(t *testing.T) {

		err = store.CreateBatch(ctx, &batch)
		require.NoError(t, err)
	})
	t.Run("get", func(t *testing.T) {

		id := batch.ID
		retrieved, err := store.GetBatch(ctx, id)
		require.NoError(t, err)
		require.Equal(t, batch, *retrieved)
	})
	t.Run("update", func(t *testing.T) {

		copy := batch
		copy.CID = batch.CID + fmt.Sprint(rand.Int())

		err = store.UpdateBatch(ctx, &copy)
		require.NoError(t, err)

		retrieved, err := store.GetBatch(ctx, copy.ID)
		require.NoError(t, err)

		require.Equal(t, copy, *retrieved)
	})
	t.Run("update status", func(t *testing.T) {

		status := rand.Int32N(11)

		err = store.UpdateBatchStatus(ctx, batch.ID, status)
		require.NoError(t, err)

		retrieved, err := store.GetBatch(ctx, batch.ID)
		require.NoError(t, err)

		require.Equal(t, status, retrieved.Status)
		// TODO: Remaining fields should be unchanged equal.
	})
	t.Run("delete", func(t *testing.T) {

		err = store.DeleteBatch(ctx, batch.ID)
		require.NoError(t, err)

		_, err := store.GetBatch(ctx, batch.ID)
		require.Error(t, err) // Should fail as the batch no longer exists
	})
}

func getDBClient(t *testing.T) *mongo.Client {
	t.Helper()

	addr := os.Getenv(MongoDBConnectionEnv)

	client, err := b7smongo.Connect(t.Context(), addr)
	require.NoError(t, err)

	return client
}
