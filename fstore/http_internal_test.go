package fstore

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/store"
	"github.com/blessnetwork/b7s/store/codec"
	"github.com/blessnetwork/b7s/testing/helpers"
	"github.com/blessnetwork/b7s/testing/mocks"
)

func TestFunction_GetJSON(t *testing.T) {

	var (
		workdir  = "/"
		manifest = mocks.GenericManifest
	)

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			payload, err := json.Marshal(manifest)
			require.NoError(t, err)
			w.Write(payload)
		}))
	defer srv.Close()

	store := store.New(helpers.InMemoryDB(t), codec.NewJSONCodec())
	fh := New(mocks.NoopLogger, store, workdir)

	var downloaded bls.FunctionManifest
	err := fh.getJSON(srv.URL, &downloaded)
	require.NoError(t, err)

	require.Equal(t, manifest, downloaded)
}

func TestFunction_GetJSONHandlesErrors(t *testing.T) {

	const (
		workdir = "/"
	)

	tests := []struct {
		name string

		statusCode int
		payload    []byte
	}{
		{
			name: "handles malformed JSON",

			statusCode: http.StatusOK,
			// JSON payload without closing brace.
			payload: []byte(`{
			"id":"generic-id",
			"name":"generic-name",
			"description":"generic-description",
			"function": {
				"id":"function-id",
				"name":"function-name",
				"runtime":"generic-runtime"
			},
			"deployment":{
				"cid":"generic-cid",
				"checksum":"1234567890",
				"uri":"generic-uri"
			},
			"runtime":{},
			"fs_root_path":"/var/tmp/bless/",
			"entry":"/var/tmp/bless/app.wasm"`), // <- missing closing brace
		},
		{
			name: "handles unexpected format",

			statusCode: http.StatusOK,
			// Valid JSON payload but wrong format - number instead of a textual field.	
			payload: []byte(`{
				"id":"generic-id",
				"name":"generic-name",
				"description":"generic-description",
				"function": {
					"id":"function-id",
					"name":"function-name",
					"runtime":"generic-runtime"
				},
				"deployment":{
					"cid":"generic-cid",
					"checksum":"1234567890",
					"uri":"generic-uri"
				},
				"runtime":{},
				"fs_root_path":"/var/tmp/bless/",
				"entry":999
			}`),
		},
		{
			name:       "handles missing data",
			statusCode: http.StatusInternalServerError,
			payload:    []byte{},
		},
	}

	for _, test := range tests {

		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(test.statusCode)
					w.Write(test.payload)
				}))
			defer srv.Close()

			fh := New(mocks.NoopLogger, newInMemoryStore(t), workdir)

			var response bls.FunctionManifest
			err := fh.getJSON(srv.URL, &response)
			require.Error(t, err)
		})
	}
}

func TestFunction_Download(t *testing.T) {

	const (
		size = 10_000
	)

	payload := getRandomPayload(t, size)

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Write(payload)
		}))
	defer srv.Close()

	workdir, err := os.MkdirTemp("", "b7s-function-download-")
	require.NoError(t, err)

	defer os.RemoveAll(workdir)

	fh := New(mocks.NoopLogger, newInMemoryStore(t), workdir)

	address := fmt.Sprintf("%s/test-file", srv.URL)
	hash := sha256.Sum256(payload)

	manifest := bls.FunctionManifest{
		Deployment: bls.Deployment{
			URI:      address,
			Checksum: fmt.Sprintf("%x", hash),
		},
	}

	path, err := fh.download(context.Background(), "", manifest)
	require.NoError(t, err)

	// Check if the file created is within the specified workdir.
	// Not the perfect way to check this, but it will do.
	require.True(t, strings.HasPrefix(path, workdir))

	downloaded, err := os.ReadFile(path)
	require.NoError(t, err)

	require.Equal(t, payload, downloaded)
}

func TestFunction_DownloadHandlesErrors(t *testing.T) {

	const (
		size = 10_000
	)
	ctx := context.Background()
	payload := getRandomPayload(t, size)

	srv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Write(payload)
		}))
	// NOTE: Server handled in a test case below.
	// Also the reason tests are not executed in parallel.

	t.Run("handles invalid checksum", func(t *testing.T) {

		workdir, err := os.MkdirTemp("", "b7s-function-download-")
		require.NoError(t, err)

		defer os.RemoveAll(workdir)

		fh := New(mocks.NoopLogger, newInMemoryStore(t), workdir)

		address := fmt.Sprintf("%s/test-file", srv.URL)
		hash := sha256.Sum256(payload)

		invalidChecksum := fmt.Sprintf("%x", hash) + "Z"

		manifest := bls.FunctionManifest{
			Deployment: bls.Deployment{
				URI:      address,
				Checksum: invalidChecksum,
			},
		}

		_, err = fh.download(ctx, "", manifest)
		require.Error(t, err)
	})
	t.Run("handles invalid URI", func(t *testing.T) {

		workdir, err := os.MkdirTemp("", "b7s-function-download-")
		require.NoError(t, err)

		defer os.RemoveAll(workdir)

		fh := New(mocks.NoopLogger, newInMemoryStore(t), workdir)

		address := fmt.Sprintf("%s/test-file", srv.URL) + "\n"
		hash := sha256.Sum256(payload)

		manifest := bls.FunctionManifest{
			Deployment: bls.Deployment{
				URI:      address,
				Checksum: fmt.Sprintf("%x", hash),
			},
		}

		_, err = fh.download(ctx, "", manifest)
		require.Error(t, err)
	})
	t.Run("handles download failure", func(t *testing.T) {

		srv.Close()

		workdir, err := os.MkdirTemp("", "b7s-function-download-")
		require.NoError(t, err)

		defer os.RemoveAll(workdir)

		fh := New(mocks.NoopLogger, newInMemoryStore(t), workdir)

		address := fmt.Sprintf("%s/test-file", srv.URL)
		hash := sha256.Sum256(payload)

		manifest := bls.FunctionManifest{
			Deployment: bls.Deployment{
				URI:      address,
				Checksum: fmt.Sprintf("%x", hash),
			},
		}

		_, err = fh.download(ctx, "", manifest)
		require.Error(t, err)
	})
}

func getRandomPayload(t *testing.T, len int) []byte {
	t.Helper()

	var seed = [32]byte([]byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ123456"))
	r := rand.NewChaCha8(seed)
	buf := make([]byte, len)
	_, err := r.Read(buf)
	require.NoError(t, err)

	return buf
}

func newInMemoryStore(t *testing.T) *store.Store {
	t.Helper()
	return store.New(helpers.InMemoryDB(t), codec.NewJSONCodec())
}
