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

func TestBatchStore_WorkItem(t *testing.T) {

	var (
		client    = getDBClient(t)
		ctx       = t.Context()
		itemCount = 10
	)

	store, err := mbs.NewBatchStore(client)
	require.NoError(t, err)

	err = store.Init(ctx)
	require.NoError(t, err)

	items := newWorkItems(t, itemCount)
	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ID
	}
	t.Run("create single items", func(t *testing.T) {
		err = store.CreateWorkItems(ctx, items[0])
		require.NoError(t, err)
	})
	t.Run("create multiple items items", func(t *testing.T) {
		err = store.CreateWorkItems(ctx, items[1:]...)
		require.NoError(t, err)
	})
	t.Run("get work items", func(t *testing.T) {

		for i, item := range items {

			id := item.ID
			retrieved, err := store.GetWorkItem(ctx, id)
			require.NoError(t, err)
			require.Equal(t, items[i], retrieved)

		}
	})
	t.Run("update", func(t *testing.T) {

		orig := items[0]
		copy := *orig
		copy.RequestID = copy.RequestID + fmt.Sprint(rand.Int32N(10))

		err = store.UpdateWorkItem(ctx, &copy)
		require.NoError(t, err)

		retrieved, err := store.GetWorkItem(ctx, copy.ID)
		require.NoError(t, err)

		require.Equal(t, copy, *retrieved)
	})
	t.Run("update status", func(t *testing.T) {

		var (
			itemID = ids[0]
			status = rand.Int32N(11)
		)

		err = store.UpdateWorkItemStatus(ctx, status, itemID)
		require.NoError(t, err)

		retrieved, err := store.GetWorkItem(ctx, itemID)
		require.NoError(t, err)

		require.Equal(t, status, retrieved.Status)
	})
	t.Run("update multiple statuses", func(t *testing.T) {

		var (
			status = rand.Int32N(11)
		)

		err = store.UpdateWorkItemStatus(ctx, status, ids...)
		require.NoError(t, err)

		for _, item := range items {

			id := item.ID
			retrieved, err := store.GetWorkItem(ctx, id)
			require.NoError(t, err)
			require.Equal(t, status, retrieved.Status)
		}
	})
	t.Run("delete items", func(t *testing.T) {

		err = store.DeleteWorkItems(ctx, ids...)
		require.NoError(t, err)

		for _, item := range items {
			_, err := store.GetWorkItem(ctx, item.ID)
			require.Error(t, err)
		}
	})
}

func newWorkItems(t *testing.T, n int) []*batchstore.WorkItemRecord {
	t.Helper()

	items := make([]*batchstore.WorkItemRecord, n)
	for i := range n {
		items[i] = &batchstore.WorkItemRecord{
			ID:        fmt.Sprintf("test.work-item-%v", rand.Int()),
			ChunkID:   fmt.Sprintf("test.chunk-%v", rand.Int()),
			RequestID: fmt.Sprintf("test-request-id-%v", rand.Int()),
			Status:    0,
		}
	}

	return items
}
