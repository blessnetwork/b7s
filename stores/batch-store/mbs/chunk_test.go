//go:build integration
// +build integration

package mbs_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"

	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"github.com/blessnetwork/b7s/stores/batch-store/mbs"
)

func TestBatchStore_Chunks(t *testing.T) {

	var (
		client      = getDBClient(t)
		ctx         = t.Context()
		recordCount = 10
	)

	store, err := mbs.NewBatchStore(client)
	require.NoError(t, err)

	err = store.Init(ctx)
	require.NoError(t, err)

	chunks := newChunks(t, recordCount)
	ids := make([]string, len(chunks))
	for i, chunk := range chunks {
		ids[i] = chunk.ID
	}

	t.Run("create single chunk", func(t *testing.T) {
		err = store.CreateChunks(ctx, chunks[0])
		require.NoError(t, err)
	})
	t.Run("create many chunks", func(t *testing.T) {
		err = store.CreateChunks(ctx, chunks[1:]...)
		require.NoError(t, err)
	})
	t.Run("get chunk", func(t *testing.T) {

		for i, chunk := range chunks {

			id := chunk.ID

			retrieved, err := store.GetChunk(ctx, id)
			require.NoError(t, err)
			require.Equal(t, chunks[i], retrieved)
		}

	})
	t.Run("update", func(t *testing.T) {

		orig := chunks[0]

		copy := *orig
		copy.RequestID = copy.RequestID + fmt.Sprint(rand.Int32N(10))

		err = store.UpdateChunk(ctx, &copy)
		require.NoError(t, err)

		retrieved, err := store.GetChunk(ctx, copy.ID)
		require.NoError(t, err)

		require.Equal(t, copy, *retrieved)
	})
	t.Run("update status", func(t *testing.T) {

		var (
			chunkID = ids[0]
			status  = rand.Int32()
		)

		err = store.UpdateChunkStatus(ctx, status, chunkID)
		require.NoError(t, err)

		// Verify change of first chunk.
		retrieved, err := store.GetChunk(ctx, chunkID)
		require.NoError(t, err)
		require.Equal(t, status, retrieved.Status)
	})
	t.Run("update multiple statuses", func(t *testing.T) {

		var (
			status = rand.Int32()
		)

		err = store.UpdateChunkStatus(ctx, status, ids...)
		require.NoError(t, err)

		for _, id := range ids {
			retrieved, err := store.GetChunk(ctx, id)
			require.NoError(t, err)
			require.Equal(t, status, retrieved.Status)
		}
	})
	t.Run("delete chunks", func(t *testing.T) {

		err = store.DeleteChunks(ctx, ids...)
		require.NoError(t, err)

		for _, chunk := range chunks {
			// Retrieving chunks should fail as they are deleted by now.
			_, err := store.GetChunk(ctx, chunk.ID)
			require.Error(t, err)
		}
	})
}

func newChunks(t *testing.T, n int) []*batchstore.ChunkRecord {

	chunks := make([]*batchstore.ChunkRecord, n)
	for i := range n {
		chunks[i] = &batchstore.ChunkRecord{
			ID:        fmt.Sprintf("test.chunk-%v", rand.Int()),
			RequestID: fmt.Sprintf("test-request-id-%v", rand.Int()),
			Status:    0,
		}
	}

	return chunks
}
