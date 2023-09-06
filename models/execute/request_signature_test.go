package execute

import (
	"testing"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/require"
)

func TestExecute_Signing(t *testing.T) {

	sampleReq := Request{
		FunctionID: "function-di",
		Method:     "method-value",
		Parameters: []Parameter{
			{
				Name:  "parameter-name",
				Value: "parameter-value",
			},
		},
		Config: Config{},
	}

	t.Run("nominal case", func(t *testing.T) {

		req := sampleReq
		priv, pub := newKey(t)

		err := req.Sign(priv)
		require.NoError(t, err)

		err = req.VerifySignature(pub)
		require.NoError(t, err)
	})
	t.Run("empty signature verification fails", func(t *testing.T) {

		req := sampleReq
		req.Signature = ""

		_, pub := newKey(t)

		err := req.VerifySignature(pub)
		require.Error(t, err)
	})
	t.Run("tampered data signature verification fails", func(t *testing.T) {

		req := sampleReq
		priv, pub := newKey(t)

		err := req.Sign(priv)
		require.NoError(t, err)

		req.FunctionID += " "

		err = req.VerifySignature(pub)
		require.Error(t, err)
	})
}

func newKey(t *testing.T) (crypto.PrivKey, crypto.PubKey) {
	t.Helper()
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	require.NoError(t, err)

	return priv, pub
}
