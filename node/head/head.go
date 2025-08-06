package head

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-metrics"

	"github.com/blessnetwork/b7s/info"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/response"
	"github.com/blessnetwork/b7s/node"
	"github.com/blessnetwork/b7s/node/internal/waitmap"
	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
)

type HeadNode struct {
	node.Core

	cfg Config

	rollCall                *rollCallQueue
	consensusResponses      *waitmap.WaitMap[string, response.FormCluster]
	workOrderResponses      *waitmap.WaitMap[string, execute.NodeResult]
	workOrderBatchResponses *waitmap.WaitMap[string, response.WorkOrderBatch]
}

func New(core node.Core, options ...Option) (*HeadNode, error) {

	// InitiaChunkResultsize config.
	cfg := DefaultConfig
	for _, option := range options {
		option(&cfg)
	}

	err := cfg.Valid()
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	head := &HeadNode{
		Core: core,
		cfg:  cfg,

		rollCall:                newQueue(rollCallQueueBufferSize),
		consensusResponses:      waitmap.New[string, response.FormCluster](0),
		workOrderResponses:      waitmap.New[string, execute.NodeResult](executionResultCacheSize),
		workOrderBatchResponses: waitmap.New[string, response.WorkOrderBatch](executionResultCacheSize),
	}

	head.Metrics().SetGaugeWithLabels(node.NodeInfoMetric, 1,
		[]metrics.Label{
			{Name: "id", Value: head.ID()},
			{Name: "version", Value: info.VcsVersion()},
			{Name: "role", Value: "head"},
		})

	return head, nil
}

func (h *HeadNode) Run(ctx context.Context) error {

	// TODO: Add a synchronous first loop or something like that so we can fail early.
	// TODO: Re-read this.
	go func(ctx context.Context) {

		// NOTE: Not a perfect solution, but the simplest one:
		// Wait a little while until some of the peers connect.
		// Else we can wait until we get some threshold N of connected peers.
		// TODO: Double check this decision.
		time.Sleep(batchResumeDelay)

		// Run first sync immediately.
		err := h.resumeUnfinishedBatches(ctx)
		if err != nil {
			h.Log().Error().
				Err(err).Msg("could not resume incomplete batches")
		}

		ticker := time.NewTicker(h.cfg.RequeueInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:

				err := h.resumeUnfinishedBatches(ctx)
				if err != nil {
					h.Log().Error().
						Err(err).
						Msg("could not resume incomplete batches")
				}

			case <-ctx.Done():
				h.Log().Info().Msg("stopping batch resume loop")
			}
		}
	}(ctx)

	return h.Core.Run(ctx, h.process)
}

func (h *HeadNode) resumeUnfinishedBatches(ctx context.Context) error {

	batches, err := h.cfg.BatchStore.FindBatches(ctx, batchstore.StatusInProgress, batchstore.StatusCreated)
	if err != nil {
		return fmt.Errorf("could not lookup incomplete batches: %w", err)
	}

	h.Log().Info().
		Int("count", len(batches)).
		Msg("found unfinished batches")

	// TODO: Decide - process batches sequentially? In parallel?
	for _, batch := range batches {

		err = h.continueBatchExecution(ctx, batch)
		if err != nil {
			h.Log().Error().
				Err(err).
				Str("batch", batch.ID).
				Msg("countinued batch execution failed")
		}
	}

	return nil
}

func newRequestID() string {
	return newUUID()
}

func newUUID() string {
	return uuid.New().String()
}
