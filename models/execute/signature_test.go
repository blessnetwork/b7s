package execute

import (
	"testing"

	"github.com/blessnetwork/b7s/models/codes"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/stretchr/testify/require"
)

func TestRequestExecute_Signing(t *testing.T) {

	sampleReq := Request{
		FunctionID: "function-di",
		Method:     "method-value",
		Arguments: []string{
			"parameter-value-1",
			"parameter-value-2",
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

func TestResultExecute_Signing(t *testing.T) {

	sampleRes := NodeResult{
		Result: Result{
			Code: codes.Unknown,
			Result: RuntimeOutput{
				Stdout:   "generic-execution-result",
				Stderr:   "generic-execution-log",
				ExitCode: 0,
			},
		},
	}

	t.Run("nominal case", func(t *testing.T) {

		res := sampleRes
		priv, pub := newKey(t)

		err := res.Sign(priv)
		require.NoError(t, err)

		err = res.VerifySignature(pub)
		require.NoError(t, err)
	})
	t.Run("empty signature verification fails", func(t *testing.T) {

		res := sampleRes
		res.Signature = ""

		_, pub := newKey(t)

		err := res.VerifySignature(pub)
		require.Error(t, err)
	})
	t.Run("tampered data signature verification fails", func(t *testing.T) {

		res := sampleRes
		priv, pub := newKey(t)

		err := res.Sign(priv)
		require.NoError(t, err)

		res.Result.Result.Stdout += " "

		err = res.VerifySignature(pub)
		require.Error(t, err)
	})
}

func newKey(t *testing.T) (crypto.PrivKey, crypto.PubKey) {
	t.Helper()
	priv, pub, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	require.NoError(t, err)

	return priv, pub
}
