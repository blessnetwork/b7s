package main

import (
	"context"
	"fmt"

	"github.com/blessnetwork/b7s/config"
	"github.com/blessnetwork/b7s/executor"
	"github.com/blessnetwork/b7s/executor/limits"
	"github.com/blessnetwork/b7s/fstore"
	"github.com/blessnetwork/b7s/models/bls"
	b7smongo "github.com/blessnetwork/b7s/mongo"
	"github.com/blessnetwork/b7s/node"
	"github.com/blessnetwork/b7s/node/head"
	"github.com/blessnetwork/b7s/node/worker"
	"github.com/blessnetwork/b7s/stores/batch-store/mbs"
)

type Node interface {
	Run(context.Context) error
}

func createWorkerNode(core node.Core, store bls.Store, cfg *config.Config) (Node, func() error, error) {

	// Create function store.
	fstore := fstore.New(log.With().Str("component", "fstore").Logger(), store, cfg.Workspace)

	// Executor options.
	execOptions := []executor.Option{
		executor.WithWorkDir(cfg.Workspace),
		executor.WithRuntimeDir(cfg.Worker.RuntimePath),
		executor.WithExecutableName(cfg.Worker.RuntimeCLI),
	}

	shutdown := func() error {
		return nil
	}
	if needLimiter(cfg) {
		limiter, err := limits.New(limits.WithCPUPercentage(cfg.Worker.CPUPercentageLimit), limits.WithMemoryKB(cfg.Worker.MemoryLimitKB))
		if err != nil {
			return nil, shutdown, fmt.Errorf("could not create resource limiter")
		}

		shutdown = func() error {
			return limiter.Shutdown()
		}

		execOptions = append(execOptions, executor.WithLimiter(limiter))
	}

	// Create an executor.
	executor, err := executor.New(log.With().Str("component", "executor").Logger(), execOptions...)
	if err != nil {
		return nil, shutdown, fmt.Errorf("could not create an executor: %w", err)
	}

	worker, err := worker.New(core, fstore, executor,
		worker.AttributeLoading(cfg.LoadAttributes),
		worker.Workspace(cfg.Workspace),
	)
	if err != nil {
		return nil, shutdown, fmt.Errorf("could not create a worker node: %w", err)
	}

	return worker, shutdown, nil
}

func createHeadNode(ctx context.Context, core node.Core, cfg *config.Config) (Node, error) {

	var opts []head.Option

	batchServer := cfg.Head.BatchDBServer
	if batchServer != "" {
		cli, err := b7smongo.Connect(ctx, batchServer)
		if err != nil {
			return nil, fmt.Errorf("could not connect to batch server: %w", err)
		}

		bs, err := mbs.NewBatchStore(cli)
		if err != nil {
			return nil, fmt.Errorf("could not create batch store: %w", err)
		}

		opts = append(opts, head.BatchStore(bs))
	}

	head, err := head.New(core, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not create a head node: %w", err)
	}

	return head, nil
}
