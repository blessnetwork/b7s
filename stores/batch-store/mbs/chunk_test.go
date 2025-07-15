//go:build integration
// +build integration

package mbs_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"github.com/blessnetwork/b7s/stores/batch-store/mbs"
)

func TestChunkStore(t *testing.T) {

	var (
		client = getDBClient(t)
		ctx    = t.Context()
	)

	store, err := mbs.NewBatchStore(client)
	require.NoError(t, err)

	err = store.Init(ctx)
	require.NoError(t, err)

	chunk := batchstore.ChunkRecord{
		ID:        uuid.NewString(),
		RequestID: "test-request-id",
		Status:    0,
	}

	t.Run("create", func(t *testing.T) {

		err = store.CreateChunk(ctx, &chunk)
		require.NoError(t, err)
	})
	t.Run("get", func(t *testing.T) {

		id := chunk.ID
		retrieved, err := store.GetChunk(ctx, id)
		require.NoError(t, err)
		require.Equal(t, chunk, *retrieved)
	})
	t.Run("update", func(t *testing.T) {

		copy := chunk
		copy.RequestID = chunk.RequestID + fmt.Sprint(rand.Int())

		err = store.UpdateChunk(ctx, &copy)
		require.NoError(t, err)

		retrieved, err := store.GetChunk(ctx, copy.ID)
		require.NoError(t, err)

		require.Equal(t, copy, *retrieved)
	})
	t.Run("update status", func(t *testing.T) {

		status := rand.Int32N(11)

		err = store.UpdateChunkStatus(ctx, chunk.ID, status)
		require.NoError(t, err)

		retrieved, err := store.GetChunk(ctx, chunk.ID)
		require.NoError(t, err)

		require.Equal(t, status, retrieved.Status)
		// TODO: Remaining fields should be unchanged equal.
	})
	t.Run("delete", func(t *testing.T) {

		err = store.DeleteChunk(ctx, chunk.ID)
		require.NoError(t, err)

		_, err := store.GetChunk(ctx, chunk.ID)
		require.Error(t, err)
	})
}
