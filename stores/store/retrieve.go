package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/bls"
)

func (s *Store) RetrievePeer(_ context.Context, id peer.ID) (bls.Peer, error) {

	idBytes, err := id.MarshalBinary()
	if err != nil {
		return bls.Peer{}, fmt.Errorf("could not serialize peer ID: %w", err)
	}

	key := encodeKey(PrefixPeer, idBytes)
	var peer bls.Peer
	err = s.retrieve(key, &peer)
	if err != nil {
		return bls.Peer{}, fmt.Errorf("could not retrieve value: %w", err)
	}

	return peer, nil
}

func (s *Store) RetrievePeers(_ context.Context) ([]bls.Peer, error) {

	peers := make([]bls.Peer, 0)

	opts := prefixIterOptions([]byte{PrefixPeer})
	it, err := s.db.NewIter(opts)
	if err != nil {
		return nil, fmt.Errorf("could not create iterator: %w", err)
	}
	for it.First(); it.Valid(); it.Next() {

		var peer bls.Peer
		err := s.retrieve(it.Key(), &peer)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve peer (key: %x): %w", it.Key(), err)
		}

		peers = append(peers, peer)
	}

	return peers, nil
}

func (s *Store) RetrieveFunction(_ context.Context, cid string) (bls.FunctionRecord, error) {

	key := encodeKey(PrefixFunction, cid)
	var function bls.FunctionRecord
	err := s.retrieve(key, &function)
	if err != nil {
		return bls.FunctionRecord{}, fmt.Errorf("could not retrieve function record: %w", err)
	}

	return function, nil
}

func (s *Store) RetrieveFunctions(_ context.Context) ([]bls.FunctionRecord, error) {

	functions := make([]bls.FunctionRecord, 0)

	opts := prefixIterOptions([]byte{PrefixFunction})
	it, err := s.db.NewIter(opts)
	if err != nil {
		return nil, fmt.Errorf("could not create iterator: %w", err)
	}
	for it.First(); it.Valid(); it.Next() {

		var function bls.FunctionRecord
		err := s.retrieve(it.Key(), &function)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve functioN (key: %x): %w", it.Key(), err)
		}

		functions = append(functions, function)
	}

	return functions, nil
}

func (s *Store) retrieve(key []byte, out any) error {

	value, closer, err := s.db.Get(key)
	if err != nil {
		if errors.Is(err, pebble.ErrNotFound) {
			return bls.ErrNotFound
		}
		return fmt.Errorf("could not retrieve value: %w", err)
	}
	// Closer must be called else a memory leak occurs.
	defer closer.Close()

	err = s.codec.Unmarshal(value, out)
	if err != nil {
		return fmt.Errorf("cold not decode record: %w", err)
	}

	return nil
}

func prefixIterOptions(prefix []byte) *pebble.IterOptions {
	return &pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: iteratorPrefixUpperBound(prefix),
	}
}

func iteratorPrefixUpperBound(prefix []byte) []byte {

	end := make([]byte, len(prefix))
	copy(end, prefix)
	for i := len(end) - 1; i >= 0; i-- {
		end[i] = end[i] + 1
		if end[i] != 0 {
			return end[:i+1]
		}
	}

	return nil
}
