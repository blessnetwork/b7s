package head

import (
	"context"
	"slices"
	"testing"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/testing/mocks"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"
)

func TestSendBatch(t *testing.T) {

	var (
		ctx   = context.Background()
		peers = slices.Clone(mocks.GenericPeerIDs[:3])

		assignments = map[peer.ID]*request.WorkOrderBatch{
			peers[0]: {},
			peers[1]: {},
			peers[2]: {},
		}

		core        = mocks.BaselineNodeCore(t)
		faultyPeers = []peer.ID{peers[0], peers[1]}
	)

	t.Run("nominal case", func(t *testing.T) {

		head, err := New(core)
		require.NoError(t, err)

		core.SendFunc = func(_ context.Context, id peer.ID, _ bls.Message) error {
			return nil
		}

		err = head.sendBatch(ctx, assignments)
		require.NoError(t, err)
	})
	t.Run("partial send  fail", func(t *testing.T) {

		head, err := New(core)
		require.NoError(t, err)

		core.SendFunc = func(_ context.Context, id peer.ID, _ bls.Message) error {

			// Simulate faulty peers failing
			if slices.Contains(faultyPeers, id) {
				return sendErr(id)
			}

			return nil
		}

		err = head.sendBatch(ctx, assignments)
		require.Error(t, err)

		var sendErr *batchSendError
		require.ErrorAs(t, err, &sendErr)

		require.Len(t, sendErr.errors, len(faultyPeers))
		for i, err := range sendErr.errors {
			require.Equal(t, faultyPeers[i], err.target)
		}
	})
}
