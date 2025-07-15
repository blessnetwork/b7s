package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/stretchr/testify/require"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/store"
	"github.com/blessnetwork/b7s/store/codec"
	"github.com/blessnetwork/b7s/testing/helpers"
	"github.com/blessnetwork/b7s/testing/mocks"
)

func TestStore_PeerOperations(t *testing.T) {
	db := helpers.InMemoryDB(t)
	defer db.Close()
	ctx := context.Background()

	peer := helpers.CreateRandomPeers(t, 1)[0]
	store := store.New(db, codec.NewJSONCodec())

	t.Run("save peer", func(t *testing.T) {
		err := store.SavePeer(ctx, peer)
		require.NoError(t, err)
	})
	t.Run("retrieve peer", func(t *testing.T) {
		retrieved, err := store.RetrievePeer(ctx, peer.ID)
		require.NoError(t, err)

		require.Equal(t, peer, retrieved)
	})
	t.Run("remove peer", func(t *testing.T) {
		err := store.RemovePeer(ctx, peer.ID)
		require.NoError(t, err)

		// Verify peer is gone.
		_, err = store.RetrievePeer(ctx, peer.ID)
		require.ErrorIs(t, err, bls.ErrNotFound)
	})
}

func TestStore_RetrievePeers(t *testing.T) {
	db := helpers.InMemoryDB(t)
	defer db.Close()
	store := store.New(db, codec.NewJSONCodec())
	ctx := context.Background()

	count := 10
	peers := make(map[peer.ID]bls.Peer)
	generated := helpers.CreateRandomPeers(t, count)
	for _, peer := range generated {
		peers[peer.ID] = peer
	}

	// Save peers.
	for _, peer := range peers {
		err := store.SavePeer(ctx, peer)
		require.NoError(t, err)
	}

	retrieved, err := store.RetrievePeers(ctx)
	require.NoError(t, err)
	require.Len(t, retrieved, count)

	// Verify peers.
	for _, peer := range retrieved {
		require.Equal(t, peers[peer.ID], peer)
	}
}

func TestStore_FunctionOperations(t *testing.T) {
	db := helpers.InMemoryDB(t)
	defer db.Close()

	function := mocks.GenericFunctionRecord
	store := store.New(db, codec.NewJSONCodec())
	ctx := context.Background()

	t.Run("save function", func(t *testing.T) {
		err := store.SaveFunction(ctx, function)
		require.NoError(t, err)
	})
	t.Run("retrieve function", func(t *testing.T) {
		retrieved, err := store.RetrieveFunction(ctx, function.CID)
		require.NoError(t, err)

		require.Equal(t, function, retrieved)
	})

	t.Run("remove function", func(t *testing.T) {
		err := store.RemoveFunction(ctx, function.CID)
		require.NoError(t, err)

		// Verify function is gone.
		_, err = store.RetrieveFunction(ctx, function.CID)
		require.ErrorIs(t, err, bls.ErrNotFound)
	})
}

func TestStore_RetrieveFunctions(t *testing.T) {
	db := helpers.InMemoryDB(t)
	defer db.Close()
	store := store.New(db, codec.NewJSONCodec())
	ctx := context.Background()

	count := 10
	functions := make(map[string]bls.FunctionRecord)
	for i := 0; i < count; i++ {

		fn := bls.FunctionRecord{
			CID:      fmt.Sprintf("dummy-cid-%v", i),
			URL:      fmt.Sprintf("https://example.com/dummy-url-%v", i),
			Manifest: mocks.GenericManifest,
			Archive:  fmt.Sprintf("/var/tmp/archive-%v.tar.gz", i),
			Files:    fmt.Sprintf("/var/tmp/files/%v", i),
		}

		functions[fn.CID] = fn
	}

	// Save functions.
	for _, fn := range functions {
		err := store.SaveFunction(ctx, fn)
		require.NoError(t, err)
	}

	retrieved, err := store.RetrieveFunctions(ctx)
	require.NoError(t, err)
	require.Len(t, retrieved, count)

	// Verify functions.
	for _, fn := range retrieved {
		require.Equal(t, functions[fn.CID], fn)
	}
}

func TestStore_HandlesFailures(t *testing.T) {

	db := helpers.InMemoryDB(t)
	defer db.Close()
	ctx := context.Background()

	t.Run("retrieving missing peer fails", func(t *testing.T) {
		store := store.New(db, codec.NewJSONCodec())

		_, err := store.RetrievePeer(ctx, mocks.GenericPeerID)
		require.Error(t, err)
	})
	t.Run("retrieving missing function fails", func(t *testing.T) {
		store := store.New(db, codec.NewJSONCodec())

		_, err := store.RetrieveFunction(ctx, mocks.GenericString)
		require.Error(t, err)
	})
	t.Run("save peer handles marshalling failures", func(t *testing.T) {

		codec := mocks.BaselineCodec(t)
		codec.MarshalFunc = func(any) ([]byte, error) {
			return nil, mocks.GenericError
		}
		store := store.New(db, codec)

		err := store.SavePeer(ctx, mocks.GenericPeer)
		require.Error(t, err)
	})
	t.Run("save function handles marshalling failures", func(t *testing.T) {

		codec := mocks.BaselineCodec(t)
		codec.MarshalFunc = func(any) ([]byte, error) {
			return nil, mocks.GenericError
		}
		store := store.New(db, codec)

		err := store.SaveFunction(ctx, mocks.GenericFunctionRecord)
		require.Error(t, err)
	})
	t.Run("retrieve peer handles unmarshalling failures", func(t *testing.T) {

		unmarshalErr := errors.New("unmarshalling error")
		codec := mocks.BaselineCodec(t)
		codec.MarshalFunc = func(obj any) ([]byte, error) {
			return json.Marshal(obj)
		}
		codec.UnmarshalFunc = func([]byte, any) error {
			return unmarshalErr
		}
		store := store.New(db, codec)

		// First, save the peer so we don't end up with a "not found" error.
		peer := helpers.CreateRandomPeers(t, 1)[0]
		err := store.SavePeer(ctx, peer)
		require.NoError(t, err)

		_, err = store.RetrievePeer(ctx, peer.ID)
		require.Error(t, err)
		require.ErrorIs(t, err, unmarshalErr)
	})
	t.Run("retrieve function handles unmarshalling failures", func(t *testing.T) {

		unmarshalErr := errors.New("unmarshalling error")
		codec := mocks.BaselineCodec(t)
		codec.MarshalFunc = func(obj any) ([]byte, error) {
			return json.Marshal(obj)
		}
		codec.UnmarshalFunc = func([]byte, any) error {
			return unmarshalErr
		}
		store := store.New(db, codec)

		// First, save the peer so we don't end up with a "not found" error.
		err := store.SaveFunction(ctx, mocks.GenericFunctionRecord)
		require.NoError(t, err)

		_, err = store.RetrieveFunction(ctx, mocks.GenericFunctionRecord.CID)
		require.Error(t, err)
		require.ErrorIs(t, err, unmarshalErr)
	})
}
