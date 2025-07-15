package worker

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-metrics"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/consensus"
	"github.com/blessnetwork/b7s/models/codes"
	"github.com/blessnetwork/b7s/models/request"
)

func (w *Worker) processRollCall(ctx context.Context, from peer.ID, req request.RollCall) error {

	w.Metrics().IncrCounterWithLabels(rollCallsSeenMetric, 1, []metrics.Label{{Name: "function", Value: req.FunctionID}})

	log := w.Log().With().
		Stringer("origin", from).
		Str("request", req.RequestID).
		Str("function", req.FunctionID).Logger()

	log.Debug().Msg("received roll call request")

	// TODO: (raft) temporary measure - at the moment we don't support multiple raft clusters on the same node at the same time.
	if req.Consensus == consensus.Raft && w.haveRaftClusters() {
		log.Warn().Msg("cannot respond to a roll call as we're already participating in one raft cluster")
		return nil
	}

	if req.Attributes != nil {

		if w.attributes == nil {
			log.Info().Msg("skipping attributed roll call - no attributes set")
			return nil
		}

		err := haveAttributes(*w.attributes, *req.Attributes)
		if err != nil {
			log.Info().Err(err).Msg("skipping attributed roll call - we do not match requested attributes")
			return nil
		}
	}

	// Check if we have this function installed.
	installed, err := w.fstore.IsInstalled(req.FunctionID)
	if err != nil {

		sendErr := w.Send(ctx, from, req.Response(codes.Error))
		if sendErr != nil {
			// Log send error but choose to return the original error.
			log.Error().Err(sendErr).Stringer("to", from).Msg("could not send response")
		}

		return fmt.Errorf("could not check if function is installed: %w", err)
	}

	// We don't have this function - install it now.
	if !installed {

		log.Info().Msg("roll call but function not installed, installing now")

		err = w.installFunction(ctx, req.FunctionID, manifestURLFromCID(req.FunctionID))
		if err != nil {
			sendErr := w.Send(ctx, from, req.Response(codes.Error))
			if sendErr != nil {
				// Log send error but choose to return the original error.
				log.Error().Err(sendErr).Stringer("to", from).Msg("could not send response")
			}
			return fmt.Errorf("could not install function: %w", err)
		}
	}

	log.Info().Msg("reporting for roll call")

	w.Metrics().IncrCounterWithLabels(rollCallsAppliedMetric, 1, []metrics.Label{{Name: "function", Value: req.FunctionID}})

	// Send positive response.
	err = w.Send(ctx, from, req.Response(codes.Accepted))
	if err != nil {
		return fmt.Errorf("could not send response: %w", err)
	}

	return nil
}

// Temporary measure - we can't have multiple Raft clusters at this point. Remove when we remove this limitation.
func (w *Worker) haveRaftClusters() bool {

	found := false
	w.clusters.WithRLock(func(data map[string]consensusExecutor) {
		for _, cluster := range data {
			if cluster.Consensus() == consensus.Raft {
				found = true
				return
			}
		}
	})

	return found
}

func manifestURLFromCID(cid string) string {
	return fmt.Sprintf("https://%s.ipfs.w3s.link/manifest.json", cid)
}
