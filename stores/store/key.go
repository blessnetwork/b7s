package store

import (
	"fmt"
)

func encodeKey(prefix uint8, segments ...any) []byte {

	key := []byte{prefix}

	for _, segment := range segments {
		switch s := segment.(type) {

		// Technically it would be nicer to have this an actual peer.ID.
		// However, peerID MarshalBinary() method returns an error, meaning we would need to
		// check it here or rely on the fact that it never errs (which it doesn't - in the current implementation).
		// Having the `encodeKey` function return an error here leads to a lot of unnecessary throughout the package.
		case []byte:

			key = append(key, Separator)
			key = append(key, s...)

		case string:
			key = append(key, Separator)
			key = append(key, []byte(s)...)

		default:
			panic(fmt.Sprintf("unsupported type (%T)", segment))
		}
	}

	return key
}
