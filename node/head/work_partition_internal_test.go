package head

import (
	"math/rand/v2"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
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
