package worker

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.opentelemetry.io/otel/trace"

	"github.com/blessnetwork/b7s/consensus"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/telemetry/tracing"
)

func (w *Worker) processWorkOrder(ctx context.Context, from peer.ID, req request.WorkOrder) error {

	w.Metrics().IncrCounterWithLabels(workOrderMetric, 1, []metrics.Label{{Name: "function", Value: req.FunctionID}})

	requestID := req.RequestID
	if requestID == "" {
		return errors.New("request ID missing")
	}

	ctx, span := w.Tracer().Start(ctx, spanWorkOrder, trace.WithAttributes(tracing.ExecutionAttributes(requestID, req.Request)...))
	defer span.End()

	log := w.Log().With().Str("request", requestID).Str("function", req.FunctionID).Logger()

	// NOTE: In case of an error, we do not return early from this function.
	// Instead, we send the response back to the caller, whatever it may be.
	code, result, err := w.execute(ctx, requestID, req.Timestamp, req.Request, from)
	if err != nil {
		log.Error().Err(err).Stringer("peer", from).Msg("execution failed")
	}

	metadata, err := w.cfg.MetadataProvider.Metadata(req.Request, result.Result)
	if err != nil {
		log.Error().Err(err).Msg("could not get metadata for the execution result")
	}

	switch code {

	case codes.NoContent:
		// There's little benefit to sending a response just to say we didn't execute anything.
		log.Info().Msg("no execution done - stopping")
		return nil

	case codes.OK:
		w.executeResponses.Set(requestID, execute.NodeResult{Result: result, Metadata: metadata})
	}

	// Prepare a work order response.
	res := req.Response(code, result).WithMetadata(metadata)

	log.Info().Stringer("code", code).Msg("execution complete")

	// Send the response, whatever it may be (success or failure).
	err = w.Send(ctx, from, res)
	if err != nil {
		return fmt.Errorf("could not send response: %w", err)
	}

	return nil
}

func (w *Worker) execute(ctx context.Context, requestID string, timestamp time.Time, req execute.Request, from peer.ID) (codes.Code, execute.Result, error) {

	// Check if we have function in store.
	functionInstalled, err := w.fstore.IsInstalled(req.FunctionID)
	if err != nil {
		return codes.Error, execute.Result{}, fmt.Errorf("could not lookup function in store: %w", err)
	}

	if !functionInstalled {
		return codes.NotFound, execute.Result{}, nil
	}

	// Determine if we should just execute this function, or are we part of the cluster.

	// Here we actually have a bit of a conceptual problem with having the same models for head and worker node.
	// Head node receives client requests so it can expect _some_ type of inaccuracy there. Worker node receives
	// execution requests from the head node, so it shouldn't really tolerate errors/ambiguities.
	cs, err := consensus.Parse(req.Config.ConsensusAlgorithm)
	if err != nil {
		return codes.Error, execute.Result{}, fmt.Errorf("could not parse consensus algorithm from the head node request, aborting (value: %s): %w", req.Config.ConsensusAlgorithm, err)
	}

	// We are not part of a cluster - just execute the request.
	if !consensusRequired(cs) {

		res, err := w.executor.ExecuteFunction(ctx, requestID, req)
		if err != nil {
			return res.Code, res, fmt.Errorf("execution failed: %w", err)
		}

		return res.Code, res, nil
	}

	// Now we KNOW we need a consensus. A cluster must already exist.
	cluster, ok := w.clusters.Get(requestID)
	if !ok {
		return codes.Error, execute.Result{}, fmt.Errorf("consensus required but no cluster found; omitted cluster formation message or error forming cluster (request: %s)", requestID)
	}

	log := w.Log().With().
		Str("request", requestID).
		Str("function", req.FunctionID).
		Stringer("consensus", cs).
		Logger()

	log.Info().Msg("work order to be executed as part of a cluster")

	code, value, err := cluster.Execute(from, requestID, timestamp, req)
	if err != nil {
		return codes.Error, execute.Result{}, fmt.Errorf("execution failed: %w", err)
	}

	log.Info().Stringer("code", code).Msg("node processed the work order")

	return code, value, nil
}
