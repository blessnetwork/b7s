//go:build integration
// +build integration

package node

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/blessnetwork/b7s/consensus"
	"github.com/blessnetwork/b7s/executor"
	"github.com/blessnetwork/b7s/fstore"
	"github.com/blessnetwork/b7s/host"
	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/node"
	"github.com/blessnetwork/b7s/node/head"
	"github.com/blessnetwork/b7s/node/worker"
	"github.com/blessnetwork/b7s/stores/store"
	"github.com/blessnetwork/b7s/stores/store/codec"
	"github.com/blessnetwork/b7s/testing/helpers"
	"github.com/blessnetwork/b7s/testing/mocks"
)

const (
	loopback = "127.0.0.1"

	startupDelay = 5 * time.Second

	cleanupDisableEnv = "B7S_INTEG_CLEANUP_DISABLE"
	runtimeDirEnv     = "B7S_INTEG_RUNTIME_DIR"
)

type runnable interface {
	Run(context.Context) error
}

type nodeScaffolding struct {
	dir     string
	db      *pebble.DB
	host    *host.Host
	logFile *os.File
	node    runnable
}

func instantiateNode(t *testing.T, nodeDir string, role bls.NodeRole) *nodeScaffolding {
	t.Helper()

	// Bootstrap node directory.
	dir, err := os.MkdirTemp("", nodeDir)
	require.NoError(t, err)

	logName := filepath.Join(dir, fmt.Sprintf("%v-log.json", role.String()))
	// Create logger.
	logFile, err := os.Create(logName)
	require.NoError(t, err)
	logger := zerolog.New(logFile)

	// Create node libp2p host.
	host, err := host.New(logger, loopback, 0)
	require.NoError(t, err)

	core := node.NewCore(logger, host)

	// If we're creating a head node - we have everything we need.
	if role == bls.HeadNode {

		headNode, err := head.New(core)
		require.NoError(t, err)

		return &nodeScaffolding{
			dir:     dir,
			logFile: logFile,
			host:    host,
			node:    headNode,
		}
	}

	var (
		dbDir      = filepath.Join(dir, "db")
		workspace  = filepath.Join(dir, "workspace")
		runtimeDir = os.Getenv(runtimeDirEnv)
	)

	// We're creating a worker node.

	// Open a DB and initialize an fstore.
	db, err := pebble.Open(dbDir, &pebble.Options{})
	require.NoError(t, err)
	fstore := fstore.New(logger, store.New(db, codec.NewJSONCodec()), workspace)

	executor, err := executor.New(logger, executor.WithRuntimeDir(runtimeDir), executor.WithWorkDir(workspace))
	require.NoError(t, err)

	worker, err := worker.New(core, fstore, executor, worker.Workspace(workspace))
	require.NoError(t, err)

	ns := nodeScaffolding{
		dir:     dir,
		db:      db,
		logFile: logFile,
		host:    host,
		node:    worker,
	}

	return &ns
}

// client is an external actor that can interact with the nodes.
type client struct {
	host *host.Host
}

func createClient(t *testing.T) *client {
	t.Helper()

	host, err := host.New(mocks.NoopLogger, loopback, 0)
	require.NoError(t, err)

	c := client{
		host: host,
	}

	return &c
}

func (c *client) sendInstallMessage(ctx context.Context, to peer.ID, manifestURL string, cid string) error {

	req := request.InstallFunction{
		ManifestURL: manifestURL,
		CID:         cid,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("could not encode message: %w", err)
	}

	err = c.host.SendMessage(ctx, to, payload)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}

func (c *client) sendExecutionMessage(ctx context.Context, to peer.ID, cid string, method string, consensus consensus.Type, count int) error {

	req := request.Execute{
		Request: execute.Request{
			FunctionID: cid,
			Method:     method,
			Config: execute.Config{
				NodeCount: count,
			},
		},
	}
	if consensus.Valid() {
		req.Config.ConsensusAlgorithm = consensus.String()
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("could not encode message: %w", err)
	}

	err = c.host.SendMessage(ctx, to, payload)
	if err != nil {
		return fmt.Errorf("could not send message: %w", err)
	}

	return nil
}

func createFunctionServer(t *testing.T, manifestPath string, deploymentPath string, archivePath string, cid string) *helpers.FunctionServer {

	manifest := bls.FunctionManifest{
		Name:            "hello",
		FSRootPath:      "./",
		DriversRootPath: "",
		LimitedFuel:     200_000_000,
		LimitedMemory:   120,
		Entry:           "hello.wasm",
	}

	return helpers.CreateFunctionServer(t, manifestPath, manifest, deploymentPath, archivePath, cid)
}

func getStreamPayload(t *testing.T, stream network.Stream, output any) {
	t.Helper()

	buf := bufio.NewReader(stream)
	payload, err := buf.ReadBytes('\n')
	require.ErrorIs(t, err, io.EOF)

	err = json.Unmarshal(payload, output)
	require.NoError(t, err)
}

func cleanupDisabled() bool {
	return os.Getenv(cleanupDisableEnv) == "yes"
}
