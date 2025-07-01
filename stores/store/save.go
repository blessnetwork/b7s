package store

import (
	"context"
	"fmt"

	"github.com/cockroachdb/pebble"

	"github.com/blessnetwork/b7s/models/bls"
)

func (s *Store) SavePeer(_ context.Context, peer bls.Peer) error {

	id, err := peer.ID.MarshalBinary()
	if err != nil {
		return fmt.Errorf("could not serialize peer ID: %w", err)
	}

	key := encodeKey(PrefixPeer, id)
	err = s.save(key, peer)
	if err != nil {
		return fmt.Errorf("could not save peer: %w", err)
	}

	return nil
}

func (s *Store) SaveFunction(_ context.Context, function bls.FunctionRecord) error {

	key := encodeKey(PrefixFunction, function.CID)
	err := s.save(key, function)
	if err != nil {
		return fmt.Errorf("could not save function: %w", err)
	}

	return nil
}

func (s *Store) save(key []byte, value any) error {

	encoded, err := s.codec.Marshal(value)
	if err != nil {
		return fmt.Errorf("could not encode value: %w", err)
	}

	err = s.db.Set(key, encoded, pebble.Sync)
	if err != nil {
		return fmt.Errorf("could not store value: %w", err)
	}

	return nil
}
