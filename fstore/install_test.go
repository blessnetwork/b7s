package fstore_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/blessnetwork/b7s/fstore"
	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/stores/store"
	"github.com/blessnetwork/b7s/stores/store/codec"
	"github.com/blessnetwork/b7s/testing/helpers"
	"github.com/blessnetwork/b7s/testing/mocks"
)

func TestFunction_Install(t *testing.T) {

	const (
		manifestURL = "manifest.json"
		functionURL = "function.tar.gz"
		testFile    = "testdata/testFunction.tar.gz"

		testCID = "dummy-cid"
	)
	ctx := context.Background()

	workdir, err := os.MkdirTemp("", "b7s-function-get-")
	require.NoError(t, err)

	defer os.RemoveAll(workdir)

	functionPayload, err := os.ReadFile(testFile)
	require.NoError(t, err)

	hash := sha256.Sum256(functionPayload)

	// We'll create two servers, so we can link one to the other.
	msrv, fsrv := createServers(t, manifestURL, functionURL, functionPayload)
	defer fsrv.Close()
	defer msrv.Close()

	fh := fstore.New(mocks.NoopLogger, newInMemoryStore(t), workdir)

	t.Run("function install works", func(t *testing.T) {

		installed, err := fh.IsInstalled(testCID)
		require.NoError(t, err)
		require.False(t, installed)

		address := fmt.Sprintf("%s/%v", msrv.URL, manifestURL)
		err = fh.Install(ctx, address, testCID)
		require.NoError(t, err)

		installed, err = fh.IsInstalled(testCID)
		require.NoError(t, err)
		require.True(t, installed)

		function, err := fh.Get(ctx, testCID)
		require.NoError(t, err)

		// Verify downloaded file.
		archive := filepath.Join(workdir, function.Manifest.Deployment.File)
		require.FileExists(t, archive)

		ok := verifyFileHash(t, archive, hash)
		require.Truef(t, ok, "file hash does not match")
	})
	t.Run("function installation info ok", func(t *testing.T) {

		ok, err := fh.IsInstalled(testCID)
		require.NoError(t, err)

		require.True(t, ok, "function installation info incorrect")
	})
	t.Run("function reported as not installed if files are missing", func(t *testing.T) {

		err = os.RemoveAll(workdir)
		require.NoError(t, err)

		ok, err := fh.IsInstalled(testCID)
		require.NoError(t, err)

		require.False(t, ok, "function installation info incorrect")
	})
}

func TestFunction_InstallHandlesErrors(t *testing.T) {

	const (
		manifestURL = "manifest.json"
		functionURL = "function.tar.gz"
		testFile    = "testdata/testFunction.tar.gz"

		testCID = "dummy-cid"
	)
	ctx := context.Background()

	functionPayload, err := os.ReadFile(testFile)
	require.NoError(t, err)

	// We'll create two servers, so we can link one to the other.
	msrv, fsrv := createServers(t, manifestURL, functionURL, functionPayload)
	// NOTE: Server shutdown handled in test cases below.

	t.Run("download ok but failure to save returns no error", func(t *testing.T) {

		workdir, err := os.MkdirTemp("", "b7s-function-get-")
		require.NoError(t, err)

		defer os.RemoveAll(workdir)

		store := mocks.BaselineStore(t)
		store.SaveFunctionFunc = func(context.Context, bls.FunctionRecord) error {
			return mocks.GenericError
		}

		fh := fstore.New(mocks.NoopLogger, store, workdir)

		address := fmt.Sprintf("%s/%v", msrv.URL, manifestURL)
		err = fh.Install(ctx, address, testCID)
		require.NoError(t, err)
	})
	t.Run("handles failure to download function", func(t *testing.T) {

		// Shutdown function server.
		fsrv.Close()

		workdir, err := os.MkdirTemp("", "b7s-function-get-")
		require.NoError(t, err)

		defer os.RemoveAll(workdir)

		fh := fstore.New(mocks.NoopLogger, newInMemoryStore(t), workdir)

		address := fmt.Sprintf("%s/%v", msrv.URL, manifestURL)
		err = fh.Install(context.Background(), address, testCID)
		require.Error(t, err)
	})
	t.Run("handles failure to fetch manifest", func(t *testing.T) {

		// Shutdown manifest server.
		msrv.Close()

		workdir, err := os.MkdirTemp("", "b7s-function-get-")
		require.NoError(t, err)

		defer os.RemoveAll(workdir)

		fh := fstore.New(mocks.NoopLogger, newInMemoryStore(t), workdir)

		address := fmt.Sprintf("%s/%v", msrv.URL, manifestURL)
		err = fh.Install(context.Background(), address, testCID)
		require.Error(t, err)
	})
}

func TestFunction_InstalledHandlesError(t *testing.T) {

	t.Run("installed handles store error", func(t *testing.T) {
		t.Parallel()

		const (
			testCID = "dummy-cid"
		)

		workdir, err := os.MkdirTemp("", "b7s-function-get-")
		require.NoError(t, err)

		defer os.RemoveAll(workdir)

		store := mocks.BaselineStore(t)
		store.RetrieveFunctionFunc = func(context.Context, string) (bls.FunctionRecord, error) {
			return bls.FunctionRecord{}, mocks.GenericError
		}

		fh := fstore.New(mocks.NoopLogger, store, workdir)

		_, err = fh.IsInstalled(testCID)
		require.Error(t, err)
	})
	t.Run("installed handles non installed function", func(t *testing.T) {
		t.Parallel()

		const (
			testCID = "dummy-cid"
		)

		workdir, err := os.MkdirTemp("", "b7s-function-get-")
		require.NoError(t, err)

		defer os.RemoveAll(workdir)

		store := mocks.BaselineStore(t)
		store.RetrieveFunctionFunc = func(context.Context, string) (bls.FunctionRecord, error) {
			return bls.FunctionRecord{}, bls.ErrNotFound
		}

		fh := fstore.New(mocks.NoopLogger, store, workdir)

		ok, err := fh.IsInstalled(testCID)
		require.NoError(t, err)
		require.False(t, ok)
	})
}

func createServers(t *testing.T, manifestURL string, functionURL string, functionPayload []byte) (manifestSrv *httptest.Server, functionSrv *httptest.Server) {
	t.Helper()

	// Create function server.
	fsrv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			path := strings.TrimPrefix(req.URL.Path, "/")
			if path != functionURL {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			w.Write(functionPayload)
		}))

	// Setup manifest that points to the function server.
	functionAddress := fmt.Sprintf("%s/%s", fsrv.URL, functionURL)
	hash := sha256.Sum256(functionPayload)
	sourceManifest := bls.FunctionManifest{
		Deployment: bls.Deployment{
			URI:      functionAddress,
			Checksum: fmt.Sprintf("%x", hash),
		},
	}

	// Create manifest server.
	msrv := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

			path := strings.TrimPrefix(req.URL.Path, "/")
			if path != manifestURL {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			payload, err := json.Marshal(sourceManifest)
			require.NoError(t, err)
			w.Write(payload)
		}))

	return msrv, fsrv
}

func verifyFileHash(t *testing.T, filename string, checksum [32]byte) bool {
	t.Helper()

	data, err := os.ReadFile(filename)
	require.NoError(t, err)

	h := sha256.Sum256(data)

	return bytes.Equal(checksum[:], h[:])
}

func newInMemoryStore(t *testing.T) *store.Store {
	t.Helper()
	return store.New(helpers.InMemoryDB(t), codec.NewJSONCodec())
}
