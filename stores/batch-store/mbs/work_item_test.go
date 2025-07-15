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

func TestWorkItemStore(t *testing.T) {

	var (
		client = getDBClient(t)
		ctx    = t.Context()
	)

	store, err := mbs.NewBatchStore(client)
	require.NoError(t, err)

	err = store.Init(ctx)
	require.NoError(t, err)

	item := batchstore.WorkItemRecord{
		ID:        uuid.NewString(),
		RequestID: "test-record",
		ChunkID:   "test-chunk",
		Status:    0,
	}

	t.Run("create", func(t *testing.T) {

		err = store.CreateWorkItem(ctx, &item)
		require.NoError(t, err)
	})
	t.Run("get", func(t *testing.T) {

		id := item.ID
		retrieved, err := store.GetWorkItem(ctx, id)
		require.NoError(t, err)
		require.Equal(t, item, *retrieved)
	})
	t.Run("update", func(t *testing.T) {

		copy := item
		copy.RequestID = item.RequestID + fmt.Sprint(rand.Int())

		err = store.UpdateWorkItem(ctx, &copy)
		require.NoError(t, err)

		retrieved, err := store.GetWorkItem(ctx, copy.ID)
		require.NoError(t, err)

		require.Equal(t, copy, *retrieved)
	})
	t.Run("update status", func(t *testing.T) {

		status := rand.Int32N(11)

		err = store.UpdateWorkItemStatus(ctx, item.ID, status)
		require.NoError(t, err)

		retrieved, err := store.GetWorkItem(ctx, item.ID)
		require.NoError(t, err)

		require.Equal(t, status, retrieved.Status)
		// TODO: Remaining fields should be unchanged equal.
	})
	t.Run("delete", func(t *testing.T) {

		err = store.DeleteWorkItem(ctx, item.ID)
		require.NoError(t, err)

		_, err := store.GetWorkItem(ctx, item.ID)
		require.Error(t, err)
	})

}
