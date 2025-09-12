package head

import (
	"math/rand/v2"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/testing/mocks"
)

func TestHead_PartitionWork(t *testing.T) {

	const (
		variants  = 20
		arglen    = 4
		peerCount = 4
	)

	var (
		peers   = mocks.GenericPeerIDs[:peerCount]
		batchID = newRequestID()
		req     = generateExecuteBatch(t, variants, arglen)
	)

	assignments := partitionWorkBatch(peers, batchID, req)

	// Each peer gets a chunk
	require.Len(t, assignments, peerCount)

	argsFound := make([][]string, 0)
	for peer, woBatch := range assignments {

		require.Contains(t, peers, peer) // Peer must be one of the specified ones.
		require.Nil(t, woBatch.Valid())  // Created work order batch must be valid.

		require.Equal(t, batchID, woBatch.RequestID)
		require.NotEmpty(t, woBatch.ChunkID)

		require.Equal(t, req.Template, woBatch.Template) // Work Order Batch must originate from our original request.

		// Accounting so we verify that we have saved all variants we put in.
		argsFound = append(argsFound, woBatch.Arguments...)

		// NOTE: This is not universally true, but for our test parameters it is:
		// input variants should be as evenly as possible split between peers.
		require.Len(t, woBatch.Arguments, variants/peerCount)
	}

	// Each argument list from the batch should produce one work item.
	// Make sure we have all of them assigned, but also have no more than we specified.
	require.ElementsMatch(t, req.Arguments, argsFound)
}

func TestHead_PartitionWorkDistribution(t *testing.T) {

	tests := []struct {
		name         string
		variants     int
		peerCount    int
		distribution map[int]int // Map chunk size to number of nodes that has chunk of that size.
	}{
		{
			name:      "even split",
			variants:  20,
			peerCount: 4,
			// Variants can be evenly split between nodes - each node should get five items.
			distribution: map[int]int{
				5: 4,
			},
		},
		{
			name:      "uneven split",
			variants:  18,
			peerCount: 5,
			// Some nodes get more than others - three nodes will get 4 items, and the rest will get 3 each.
			distribution: map[int]int{
				4: 3,
				3: 2,
			},
		},
		{
			name:      "more workers than items",
			variants:  3,
			peerCount: 10,
			// Some workers don't get anything.
			distribution: map[int]int{
				1: 3,
				0: 7,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var (
				peers = mocks.GenericPeerIDs[:test.peerCount]
				req   = generateExecuteBatch(t, test.variants, 2)
			)

			assignments := partitionWorkBatch(peers, newRequestID(), req)

			counts := make(map[int]int)
			for peer, woBatch := range assignments {
				t.Logf("peer: %v items: %v", peer.String(), len(woBatch.Arguments))
				counts[len(woBatch.Arguments)]++
			}

			require.Equal(t, test.distribution, counts)
		})
	}
}

func TestHead_PartitionWorkHandlesErrors(t *testing.T) {

	t.Run("empty peer list", func(t *testing.T) {

		req := generateExecuteBatch(t, 1, 2)

		assignments := partitionWorkBatch([]peer.ID{}, "dummy-request-id", req)
		require.Empty(t, assignments)
	})
}

func generateExecuteBatch(t *testing.T, itemCount int, arglen int) request.ExecuteBatch {
	t.Helper()

	req := request.ExecuteBatch{
		Template: request.ExecutionRequestTemplate{
			FunctionID: mocks.GenericFunctionID,
			Method:     mocks.GenericFunctionMethod,
			Config: execute.Config{
				NodeCount: rand.Int(),
				Environment: []execute.EnvVar{
					{Name: "env_var1", Value: "val1"},
					{Name: "env_var2", Value: "val2"},
				},
			},
		},
		MaxAttempts: rand.Uint32(),
	}

	variants := make([][]string, itemCount)
	for i := range variants {
		variants[i] = make([]string, arglen)
		gofakeit.Slice(&variants[i])
	}

	req.Arguments = variants

	return req
}
