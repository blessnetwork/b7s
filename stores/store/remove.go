package store

import (
	"context"
	"fmt"

	"github.com/cockroachdb/pebble"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (s *Store) RemovePeer(_ context.Context, id peer.ID) error {

	idBytes, err := id.MarshalBinary()
	if err != nil {
		return fmt.Errorf("could not encode peer ID: %w", err)
	}

	key := encodeKey(PrefixPeer, idBytes)
	err = s.remove(key)
	if err != nil {
		return fmt.Errorf("could not remove peer: %w", err)
	}

	return nil
}

func (s *Store) RemoveFunction(_ context.Context, cid string) error {

	key := encodeKey(PrefixFunction, cid)
	err := s.remove(key)
	if err != nil {
		return fmt.Errorf("could not remove function: %w", err)
	}

	return nil
}

func (s *Store) remove(key []byte) error {
	return s.db.Delete(key, pebble.Sync)
}
