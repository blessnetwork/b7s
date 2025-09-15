package head

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"github.com/blessnetwork/b7s/testing/mocks"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

func TestHead_SaveBatch(t *testing.T) {

	var (
		head = createHeadNode(t)
		bs   = mocks.BaselineBatchStore(t)

		batchID = newRequestID()
		req     = generateExecuteBatch(t, 4, 3)
	)

	head.cfg.BatchStore = bs

	bs.CreateBatchFunc = func(_ context.Context, er *batchstore.ExecuteBatchRecord) error {
		require.NotNil(t, er)
		require.Equal(t, batchID, er.ID)
		require.Equal(t, req.Template.FunctionID, er.CID)
		require.Equal(t, req.Template.Method, er.Method)
		require.Equal(t, req.Template.Config, er.Config)
		require.Equal(t, req.MaxAttempts, er.MaxAttempts)

		return nil
	}

	bs.CreateWorkItemsFunc = func(_ context.Context, items ...*batchstore.WorkItemRecord) error {

		require.Len(t, items, len(req.Arguments))
		for i, item := range items {
			require.Equal(t, batchID, item.BatchID)
			require.Empty(t, item.ChunkID)

			require.Equal(t, req.Arguments[i], item.Arguments)
			require.Equal(t, batchstore.StatusCreated, int(item.Status))
			require.Empty(t, item.Output)
			require.Zero(t, item.Attempts)
		}

		return nil
	}

	err := head.saveBatch(batchID, req)
	require.NoError(t, err)
}

func TestHead_SaveChunk(t *testing.T) {

	var (
		head = createHeadNode(t)
		bs   = mocks.BaselineBatchStore(t)

		batchID = newRequestID()

		template = request.ExecutionRequestTemplate{
			FunctionID: mocks.GenericFunctionID,
			Method:     mocks.GenericFunctionMethod,
		}

		parts = []*request.WorkOrderBatch{
			{
				Template:  template,
				RequestID: batchID,
				ChunkID:   newRequestID(),
				Arguments: [][]string{
					{"10", "11", "12"},
					{"20", "21", "22"},
					{"30", "31", "32"},
				},
			},
			{
				Template:  template,
				RequestID: batchID,
				ChunkID:   newRequestID(),
				Arguments: [][]string{
					{"40", "41", "42"},
					{"50", "51", "52"},
					{"60", "61", "62"},
				},
			},
		}

		assignments = map[peer.ID]*request.WorkOrderBatch{
			mocks.GenericPeerIDs[0]: parts[0],
			mocks.GenericPeerIDs[1]: parts[1],
		}
	)

	head.cfg.BatchStore = bs

	// Enable lookup for easier verification.
	lookup := map[string]struct {
		worker peer.ID
		wo     *request.WorkOrderBatch
	}{
		parts[0].ChunkID: {
			worker: mocks.GenericPeerIDs[0],
			wo:     parts[0],
		},
		parts[1].ChunkID: {
			worker: mocks.GenericPeerIDs[1],
			wo:     parts[1],
		},
	}

	bs.CreateChunksFunc = func(_ context.Context, chunks ...*batchstore.ChunkRecord) error {

		// Require we have precisely the number of chunks we expect.
		require.Len(t, chunks, len(parts))
		for _, chunk := range chunks {

			// We MUST have this exact chunk.
			orig, ok := lookup[chunk.ID]
			require.True(t, ok)

			require.Equal(t, batchID, chunk.BatchID)
			require.Equal(t, orig.wo.ChunkID, chunk.ID)
			require.Equal(t, orig.worker.String(), chunk.Worker)
			require.Equal(t, batchstore.StatusCreated, int(chunk.Status))
		}

		return nil
	}

	// For each chunk we should assign the work items.
	assignmentsDone := 0
	bs.AssignWorkItemsFunc = func(_ context.Context, chunkID string, ids ...string) error {
		assignmentsDone += 1

		orig, ok := lookup[chunkID]
		require.True(t, ok)
		require.Len(t, ids, len(orig.wo.Arguments))

		// Each list of arguments will produce one work item.
		var expectedIDs []string
		for _, args := range orig.wo.Arguments {
			expectedIDs = append(expectedIDs,
				workItemID(batchID, string(execute.ExecutionID(orig.wo.Template.FunctionID, orig.wo.Template.Method, args))))
		}

		// We need these sorted so comparison returns success.
		slices.Sort(ids)
		slices.Sort(expectedIDs)
		require.Equal(t, expectedIDs, ids)

		// Verify ID is in format <batch-id>/<work-item-id>
		for _, id := range ids {
			fields := strings.Split(id, "/")
			require.Len(t, fields, 2)
			require.Equal(t, batchID, fields[0])
		}

		return nil
	}

	err := head.saveChunkInfo(batchID, assignments)
	require.NoError(t, err)

	require.Equal(t, len(parts), assignmentsDone)
}
