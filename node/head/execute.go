package head

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.opentelemetry.io/otel/trace"

	cons "github.com/blessnetwork/b7s/consensus"
	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/execute"
	"github.com/blessnetwork/b7s/models/request"
	"github.com/blessnetwork/b7s/models/response"
	"github.com/blessnetwork/b7s/telemetry/tracing"
)

// NOTE: head node typically receives execution requests from the REST API. This message handling is not cognizant of subgroups.
func (h *HeadNode) processExecute(ctx context.Context, from peer.ID, req request.Execute) error {

	err := req.Valid()
	if err != nil {
		err = h.Send(ctx, from, req.Response(codes.Invalid, "").WithErrorMessage(err))
		if err != nil {
			return fmt.Errorf("could not send response: %w", err)
		}
		return nil
	}

	requestID := newRequestID()

	log := h.Log().With().
		Stringer("peer", from).
		Str("request", requestID).
		Str("function", req.FunctionID).Logger()

	if req.Config.NodeCount == 0 {
		req.Config.NodeCount = -1
	}

	code, results, cluster, err := h.execute(ctx, requestID, req)
	if err != nil {
		log.Error().Err(err).Msg("execution failed")
	}

	log.Info().Stringer("code", code).Msg("execution complete")

	res := req.Response(code, requestID).WithResults(results).WithCluster(cluster)
	// Communicate the reason for failure in these cases.
	if errors.Is(err, bls.ErrRollCallTimeout) || errors.Is(err, bls.ErrExecutionNotEnoughNodes) {
		res.ErrorMessage = err.Error()
	}

	// Send the response, whatever it may be (success or failure).
	err = h.Send(ctx, from, res)
	if err != nil {
		return fmt.Errorf("could not send response: %w", err)
	}

	return nil
}

// headExecute is called on the head node. The head node will publish a roll call and delegate an execution request to chosen nodes.
// The returned map contains execution results, mapped to the peer IDs of peers who reported them.
func (h *HeadNode) execute(ctx context.Context, requestID string, req request.Execute) (codes.Code, execute.ResultMap, execute.Cluster, error) {

	h.Metrics().IncrCounterWithLabels(executionsMetric, 1,
		[]metrics.Label{
			{Name: "function", Value: req.FunctionID},
			{Name: "consensus", Value: req.Config.ConsensusAlgorithm},
		})

	ctx, span := h.Tracer().Start(ctx, spanExecute,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(tracing.ExecutionAttributes(requestID, req.Request)...))
	defer span.End()

	// Create a logger with relevant context.
	log := h.Log().With().
		Str("request", requestID).
		Str("function", req.FunctionID).
		Int("node_count", req.Config.NodeCount).
		Logger()

	consensus, err := cons.Parse(req.Config.ConsensusAlgorithm)
	if err != nil {
		log.Error().
			Err(err).
			Str("value", req.Config.ConsensusAlgorithm).
			Stringer("default", h.cfg.DefaultConsensus).
			Msg("could not parse consensus algorithm from the user request, using default")

		consensus = h.cfg.DefaultConsensus
	}

	if consensusRequired(consensus) {
		log = log.With().Stringer("consensus", consensus).Logger()
	}

	log.Info().Msg("processing execution request")

	// Phase 1. - Issue roll call to nodes.

	// Limit for how long we wait for responses.
	t := cmp.Or(
		time.Duration(req.Config.Timeout)*time.Second,
		h.cfg.RollCallTimeout,
	)
	tctx, cancel := context.WithTimeout(ctx, t)
	defer cancel()

	reportingPeers, err := h.executeRollCall(tctx, req.RollCall(requestID, consensus), req.Topic, req.Config.NodeCount)
	if err != nil {
		code := codes.Error
		if errors.Is(err, bls.ErrRollCallTimeout) {
			code = codes.Timeout
		}

		return code, nil, execute.Cluster{}, fmt.Errorf("could not roll call peers (request: %s): %w", requestID, err)
	}

	cluster := execute.Cluster{
		Peers: reportingPeers,
	}

	// Phase 2. - Request cluster formation, if we need consensus.
	if consensusRequired(consensus) {

		log.Info().Strs("peers", bls.PeerIDsToStr(reportingPeers)).Msg("requesting cluster formation from peers who reported for roll call")

		err := h.formCluster(ctx, requestID, reportingPeers, consensus)
		if err != nil {
			return codes.Error, nil, execute.Cluster{}, fmt.Errorf("could not form cluster (request: %s): %w", requestID, err)
		}

		// When we're done, send a message to disband the cluster.
		// NOTE: We could schedule this on the worker nodes when receiving the execution request.
		// One variant I tried is waiting on the execution to be done on the leader (using a timed wait on the execution response) and starting raft shutdown after.
		// However, this can happen too fast and the execution request might not have been propagated to all of the nodes in the cluster, but "only" to a majority.
		// Doing this here allows for more wiggle room and ~probably~ all nodes will have seen the request so far.
		defer h.disbandCluster(requestID, reportingPeers)
	}

	// Phase 3. - Request execution.

	// Send the work order to peers in the cluster. Non-leaders will drop the request.
	workOrder := req.WorkOrder(requestID)

	// If we're working with PBFT, sign the request.
	if consensus == cons.PBFT {
		err := workOrder.Request.Sign(h.Host().PrivateKey())
		if err != nil {
			return codes.Error, nil, cluster, fmt.Errorf("could not sign execution request (function: %s, request: %s): %w", req.FunctionID, requestID, err)
		}
	}

	err = h.SendToMany(ctx,
		reportingPeers,
		workOrder,
		consensusRequired(consensus), // If we're using consensus, try to reach all peers.
	)
	if err != nil {
		return codes.Error, nil, cluster, fmt.Errorf("could not send execution request to peers (function: %s, request: %s): %w", req.FunctionID, requestID, err)
	}

	log.Debug().Msg("waiting for execution responses")

	var results execute.ResultMap
	if consensus == cons.PBFT {
		results = h.gatherExecutionResultsPBFT(ctx, requestID, reportingPeers)

		log.Info().Msg("received PBFT execution responses")

		retcode := codes.OK
		// Use the return code from the execution as the return code.
		for _, res := range results {
			retcode = res.Code
			break
		}

		return retcode, results, cluster, nil
	}

	results = h.gatherExecutionResults(ctx, requestID, reportingPeers)

	log.Info().Int("cluster_size", len(reportingPeers)).Int("responded", len(results)).Msg("received execution responses")

	// How many results do we have, and how many do we expect.
	respondRatio := float64(len(results)) / float64(len(reportingPeers))
	threshold := determineThreshold(req.Request)

	retcode := codes.OK
	if respondRatio == 0 {
		retcode = codes.NoContent
	} else if respondRatio < threshold {
		log.Warn().Float64("expected", threshold).Float64("have", respondRatio).Msg("threshold condition not met")
		retcode = codes.PartialContent
	}

	return retcode, results, cluster, nil
}

func (h *HeadNode) processWorkOrderResponse(ctx context.Context, from peer.ID, res response.WorkOrder) error {

	h.Log().Debug().
		Stringer("from", from).
		Str("request", res.RequestID).
		Msg("received work order response")

	key := peerRequestKey(res.RequestID, from)
	h.workOrderResponses.Set(key, res.Result)

	return nil
}

func (h *HeadNode) processWorkOrderBatchResponse(ctx context.Context, from peer.ID, res response.WorkOrderBatch) error {

	h.Log().Debug().
		Stringer("from", from).
		Str("request", res.RequestID).
		Str("chunk", res.ChunkID).
		Msg("received work order batch response")

	key := peerChunkKey(res.RequestID, res.ChunkID, from)
	h.workOrderBatchResponses.Set(key, res)

	return nil
}

func peerChunkKey(requestID string, _ string, peer peer.ID) string {
	return requestID + "/" + peer.String()
}

func determineThreshold(req execute.Request) float64 {

	if req.Config.Threshold > 0 && req.Config.Threshold <= 1 {
		return req.Config.Threshold
	}

	return defaultExecutionThreshold
}

func peerRequestKey(requestID string, peer peer.ID) string {
	return requestID + "/" + peer.String()
}

// helper function just for the sake of readibility.
func consensusRequired(c cons.Type) bool {
	return c != 0
}
