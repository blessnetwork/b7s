package worker

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-metrics"

	"github.com/blessnetwork/b7s/info"
	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/node"
	"github.com/blessnetwork/b7s/node/internal/syncmap"
	"github.com/blessnetwork/b7s/node/internal/waitmap"
	"github.com/blocklessnetwork/b7s-attributes/attributes"
)

type Worker struct {
	node.Core

	cfg Config

	executor bls.Executor
	fstore   FStore

	attributes *attributes.Attestation

	clusters         *syncmap.Map[string, consensusExecutor] // clusters maps request ID to the cluster the node belongs to.
	executeResponses *waitmap.WaitMap[string, execute.NodeResult]
}

func New(core node.Core, fstore FStore, executor bls.Executor, options ...Option) (*Worker, error) {

	// Initialize config.
	cfg := DefaultConfig
	for _, option := range options {
		option(&cfg)
	}

	err := cfg.Valid()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	worker := &Worker{
		Core: core,
		cfg:  cfg,

		fstore:           fstore,
		executor:         executor,
		clusters:         syncmap.New[string, consensusExecutor](),
		executeResponses: waitmap.New[string, execute.NodeResult](1000),
	}

	if cfg.LoadAttributes {

		attributes, err := loadAttributes(core.Host().PublicKey())
		if err != nil {
			return nil, fmt.Errorf("could not load attribute data: %w", err)
		}

		core.Log().Info().
			Any("attributes", attributes).
			Msg("node loaded attributes")

		worker.attributes = &attributes
	}

	worker.Metrics().SetGaugeWithLabels(node.NodeInfoMetric, 1,
		[]metrics.Label{
			{Name: "id", Value: worker.ID()},
			{Name: "version", Value: info.VcsVersion()},
			{Name: "role", Value: "worker"},
		})

	return worker, nil
}

func (w *Worker) Run(ctx context.Context) error {

	// Sync functions now in case they were removed from the storage.
	err := w.fstore.Sync(ctx, false)
	if err != nil {
		return fmt.Errorf("could not sync functions: %w", err)
	}

	// Start the function sync in the background to periodically check functions.
	go w.runSyncLoop(ctx)

	return w.Core.Run(ctx, w.process)
}
