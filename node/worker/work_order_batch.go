package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/models/response"
)

// TODO: Perhaps move this and keep it in a single place.
type ChunkResult struct {
	FunctionInvocation string
	Arguments          []string
	Result             execute.Result
	Metadata           any
}

func (w *Worker) processWorkOrderBatch(ctx context.Context, from peer.ID, req request.WorkOrderBatch) error {

	requestID := req.RequestID
	chunkID := req.ChunkID

	log := w.Log().With().
		Str("request", requestID).
		Str("chunk", chunkID).
		Str("function", req.Template.FunctionID).
		Logger()

	log.Info().
		Int("variants", len(req.Arguments)).
		Uint("concurrency", req.ConcurrencyLimit).
		Msg("received a batch work order")

	// TODO: Handle parallelism

	results := make(map[execute.RequestHash]*response.BatchFunctionResult)

	for _, args := range req.Arguments {

		// TODO: Fill this in.
		er := execute.Request{
			FunctionID: req.Template.FunctionID,
			Method:     req.Template.Method,
			Config:     req.Template.Config,
			Arguments:  args,
		}
		_, result, err := w.execute(ctx, req.ChunkID, time.Now(), er, from)
		if err != nil {
			log.Error().Err(err).Stringer("peer", from).Msg("execution failed")
		}

		metadata, err := w.cfg.MetadataProvider.Metadata(er, result.Result)
		if err != nil {
			log.Error().Err(err).Msg("could not get metadata from the execution result")
		}

		chunkID := er.GetExecutionID()
		results[chunkID] = &response.BatchFunctionResult{
			FunctionInvocation: execute.FunctionInvocation(er.FunctionID, er.Method),
			Arguments:          args,
			NodeResult: execute.NodeResult{
				Result:   result,
				Metadata: metadata,
			},
		}
	}

	res := response.WorkOrderBatch{
		RequestID: req.RequestID,
		ChunkID:   req.ChunkID,
		Results:   results,
	}
	err := w.Send(ctx, from, res)
	if err != nil {
		return fmt.Errorf("could not send response: %w", err)
	}

	return nil
}
