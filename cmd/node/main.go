package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"slices"

	"github.com/cockroachdb/pebble"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho/v3"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/blessnetwork/b7s/api"
	"github.com/blessnetwork/b7s/config"
	b7shost "github.com/blessnetwork/b7s/host"
	b7slog "github.com/blessnetwork/b7s/log"
	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/node"
	"github.com/blessnetwork/b7s/stores/store"
	"github.com/blessnetwork/b7s/stores/store/codec"
	"github.com/blessnetwork/b7s/stores/store/traceable"
	"github.com/blessnetwork/b7s/telemetry"
)

const (
	defaultLogLevel = zerolog.DebugLevel
)

var (
	log = zerolog.New(os.Stdout).With().Timestamp().Logger().Level(defaultLogLevel)
)

const (
	success = 0
	failure = 1
)

func main() {
	os.Exit(run())
}

func run() int {

	// Parse CLI flags and validate that the configuration is valid.
	cfg, err := config.Load()
	if err != nil {
		log.Error().Err(err).Msg("could not read configuration")
		return failure
	}

	// Update log level to what's in the config.
	log = log.Level(parseLogLevel(cfg.Log.Level))
	b7slog.DefaultLogger = log

	var (
		nodeID  string
		nodeDir string

		nodeRole = parseNodeRole(cfg.Role)

		// HTTP server will be created in two scenarios:
		// - node is a head node (head node always has a REST API)
		// - node has prometheus metrics enabled
		needHTTPServer = nodeRole == bls.HeadNode || cfg.Telemetry.Metrics.Enable
		server         *echo.Echo

		// If we have a REST API address, serve metrics there.
		serverAddress = cmp.Or(cfg.Head.RestAPI, cfg.Telemetry.Metrics.PrometheusAddress)
	)

	// Create the main context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if needHTTPServer {

		if serverAddress == "" {
			log.Error().Err(err).Msg("HTTP server address is required")
			return failure
		}

		server = createEchoServer(log)
	}

	// TODO: Change how node starts up with regards to key/no-key.
	if cfg.Connectivity.PrivateKey != "" {
		nodeID, err = peerIDFromKey(cfg.Connectivity.PrivateKey)
		if err != nil {
			log.Error().Err(err).Str("key", cfg.Connectivity.PrivateKey).Msg("could not read private key")
			return failure
		}
	}

	if cfg.Telemetry.Tracing.Enable {

		opts := []telemetry.TraceOption{
			telemetry.WithID(nodeID),
			telemetry.WithNodeRole(nodeRole),
			telemetry.WithBatchTraceTimeout(cfg.Telemetry.Tracing.ExporterBatchTimeout),
			telemetry.WithGRPCTracing(cfg.Telemetry.Tracing.GRPC.Endpoint),
			telemetry.WithHTTPTracing(cfg.Telemetry.Tracing.HTTP.Endpoint),
		}

		shutdown, err := telemetry.InitializeTracing(ctx, log.With().Str("component", "telemetry").Logger(), opts...)
		if err != nil {
			log.Error().Err(err).Msg("could not initialize tracing")
			return failure
		}
		defer func() {
			err := shutdown(ctx)
			if err != nil {
				log.Error().Err(err).Msg("could not shutdown tracing")
			}
		}()

		log.Info().Msg("tracing enabled")
	}

	if cfg.Telemetry.Metrics.Enable {

		metrics, err := telemetry.InitializeMetrics(
			telemetry.WithCounters(metricCounters()),
			telemetry.WithSummaries(metricSummaries()),
			telemetry.WithGauges(metricGauges()),
		)
		if err != nil {
			log.Error().Err(err).Msg("could not initialize metrics")
			return failure
		}
		defer metrics.Shutdown()

		log.Info().Msg("metrics enabled")

		// Setup metrics endpoint.
		server.GET("/metrics", echo.WrapHandler(telemetry.GetMetricsHTTPHandler()))

		// Echo (HTTP server) metrics.
		server.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{}))
	}

	// If we have a key, use path that corresponds to that key e.g. `.b7s_<peer-id>`.
	if nodeID != "" {
		nodeDir = generateNodeDirName(nodeID)
	} else {
		nodeDir, err = os.MkdirTemp("", ".b7s_*")
		if err != nil {
			log.Error().Err(err).Msg("could not create node directory")
			return failure
		}
	}

	// Set relevant working paths for workspace and DB.
	// If paths were set using the CLI flags, use those. Else, use generated path, e.g. .b7s_<peer-id>/<default-option-for-directory>.
	updateDirPaths(nodeDir, cfg)

	log.Info().
		Str("workspace", cfg.Workspace).
		Str("db", cfg.DB).
		Msg("filepaths used by the node")

	// Convert workspace path to an absolute one.
	workspace, err := filepath.Abs(cfg.Workspace)
	if err != nil {
		log.Error().Err(err).Str("path", cfg.Workspace).Msg("could not determine absolute path for workspace")
		return failure
	}
	cfg.Workspace = workspace

	// Open the pebble peer database.
	db, err := pebble.Open(cfg.DB, &pebble.Options{Logger: &pebbleNoopLogger{}})
	if err != nil {
		log.Error().Err(err).Str("db", cfg.DB).Msg("could not open pebble database")
		return failure
	}
	defer db.Close()

	// Create a new store.
	store := traceable.New(store.New(db, codec.NewJSONCodec()))

	// Create host.
	var dialbackPeers []bls.Peer
	if !cfg.Connectivity.NoDialbackPeers {
		dialbackPeers, err = store.RetrievePeers(ctx)
		if err != nil {
			log.Error().Err(err).Msg("could not get list of dial-back peers")
			return failure
		}
	}

	host, err := createHost(log.With().Str("component", "host").Logger(), *cfg, nodeRole, dialbackPeers...)
	if err != nil {
		log.Error().Err(err).Msg("could not create host")
		return failure
	}
	defer host.Close()

	host.Network().Notify(b7shost.NewNotifee(
		log.With().Str("component", "notifiee").Logger(),
		store,
	))

	log.Info().
		Str("id", host.ID().String()).
		Strs("addresses", host.Addresses()).
		Strs("boot_nodes", cfg.BootNodes).
		Msg("created host")

	// Ensure default topic is included in the topic list.
	if !slices.Contains(cfg.Topics, bls.DefaultTopic) {
		cfg.Topics = append(cfg.Topics, bls.DefaultTopic)
	}

	// Instantiate node.

	// First, initialize the node core, common for both node types.
	core := node.NewCore(
		log.With().Str("component", "node").Logger(),
		host,
		node.Concurrency(cfg.Concurrency),
		node.Topics(cfg.Topics),
	)

	var (
		node         Node
		nodeshutdown func() error
	)

	switch nodeRole {
	case bls.WorkerNode:
		node, nodeshutdown, err = createWorkerNode(core, store, cfg)

		if nodeshutdown != nil {
			defer func() {
				err = nodeshutdown()
				if err != nil {
					log.Error().Err(err).Msg("node shutdown function failed")
				}
			}()
		}

	case bls.HeadNode:
		node, err = createHeadNode(ctx, core, cfg)
	}
	if err != nil {
		log.Error().Err(err).Msg("could not create node")
		return failure
	}

	done := make(chan struct{})
	failed := make(chan struct{})

	// Start node main loop in a separate goroutine.
	go func() {

		log.Info().
			Stringer("role", nodeRole).
			Msg("Bless Node starting")

		err := node.Run(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Bless Node failed")
			close(failed)
		} else {
			close(done)
		}

		log.Info().Msg("Bless Node stopped")
	}()

	// Start the HTTP server if needed.
	if needHTTPServer {

		// Create an API handler if we're a head node.
		if nodeRole == bls.HeadNode {

			headNode, ok := any(node).(api.Node)
			if !ok {
				log.Error().Msg("invalid node type - not a head node")
			}

			apiHandler := api.New(log.With().Str("component", "api").Logger(), headNode)
			api.RegisterHandlers(server, apiHandler)
		}

		// Start server in a separate goroutine.
		go func() {

			log.Info().Str("address", serverAddress).Msg("HTTP server starting")

			err := server.Start(serverAddress)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Warn().Err(err).Msg("HTTP server failed")
				close(failed)
			} else {
				close(done)
			}

			log.Info().Msg("HTTP server stopped")
		}()
	}

	// Signal catching for clean shutdown.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	select {
	case <-sig:
		log.Info().Msg("Bless Node stopping")
	case <-done:
		log.Info().Msg("Bless Node done")
	case <-failed:
		log.Info().Msg("Bless Node aborted")
		return failure
	}

	// If we receive a second interrupt signal, exit immediately.
	go func() {
		<-sig
		log.Warn().Msg("forcing exit")
		os.Exit(1)
	}()

	return success
}

func createEchoServer(log zerolog.Logger) *echo.Echo {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	elog := lecho.From(log)
	server.Logger = elog
	server.Use(otelecho.Middleware(""))
	server.Use(lecho.Middleware(lecho.Config{Logger: elog}))

	return server
}

func needLimiter(cfg *config.Config) bool {
	return (cfg.Worker.CPUPercentageLimit > 0 && cfg.Worker.CPUPercentageLimit < 1.0) || cfg.Worker.MemoryLimitKB > 0
}

func updateDirPaths(root string, cfg *config.Config) {

	workspace := cfg.Workspace
	if workspace == "" {
		workspace = filepath.Join(root, config.DefaultWorkspaceName)
	}
	cfg.Workspace = workspace

	db := cfg.DB
	if db == "" {
		db = filepath.Join(root, config.DefaultDBName)
	}
	cfg.DB = db
}

func generateNodeDirName(id string) string {
	return fmt.Sprintf(".b7s_%s", id)
}

func parseLogLevel(s string) zerolog.Level {

	level, err := zerolog.ParseLevel(s)
	if err != nil {
		log.Error().Err(err).Str("level", s).Msg("could not parse log level")
		return defaultLogLevel
	}

	return level
}
