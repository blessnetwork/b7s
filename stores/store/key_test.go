package store

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/blessnetwork/b7s/testing/mocks"
)

func TestStore_KeyEncoding(t *testing.T) {

	t.Run("encoding peer ID works", func(t *testing.T) {

		idBytes, err := mocks.GenericPeerID.MarshalBinary()
		require.NoError(t, err)

		encodedKey := bytes.Join([][]byte{
			{0x1},    // peer prefix
			idBytes}, // peer ID
			[]byte{Separator}) // join segments by a separator

		key := encodeKey(PrefixPeer, idBytes)
		require.Equal(t, encodedKey, key)
	})
	t.Run("encoding function ID works", func(t *testing.T) {

		id := mocks.GenericString

		encodedKey := bytes.Join([][]byte{
			{0x2},       // function prefix
			[]byte(id)}, // function ID
			[]byte{Separator}) // join segments by a separator

		key := encodeKey(PrefixFunction, id)
		require.Equal(t, encodedKey, key)
	})
	t.Run("unsupported key type fails", func(t *testing.T) {

		require.Panics(t, func() {
			var empty struct{}
			_ = encodeKey(PrefixPeer, empty)
		})
	})

}
