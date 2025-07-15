package raft

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/hashicorp/raft"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"

	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/models/execute"
)

type FSMLogEntry struct {
	RequestID string          `json:"request_id,omitempty"`
	Origin    peer.ID         `json:"origin,omitempty"`
	Execute   execute.Request `json:"execute,omitempty"`
}

type FSMProcessFunc func(req FSMLogEntry, res execute.NodeResult)

type fsmExecutor struct {
	log        zerolog.Logger
	executor   bls.Executor
	processors []FSMProcessFunc
}

func newFsmExecutor(log zerolog.Logger, executor bls.Executor, processors ...FSMProcessFunc) *fsmExecutor {

	ps := make([]FSMProcessFunc, 0, len(processors))
	ps = append(ps, processors...)

	start := time.Now()
	ps = append(ps, func(req FSMLogEntry, _ execute.NodeResult) {
		// Global metrics handle.
		metrics.MeasureSinceWithLabels(raftExecutionTimeMetric, start, []metrics.Label{{Name: "function", Value: req.Execute.FunctionID}})
	})

	fsm := fsmExecutor{
		log:        log.With().Str("module", "fsm").Logger(),
		executor:   executor,
		processors: ps,
	}

	return &fsm
}

func (f fsmExecutor) Apply(log *raft.Log) any {

	f.log.Info().Msg("applying log entry")

	// Unpack the execution request.
	payload := log.Data

	var logEntry FSMLogEntry
	err := json.Unmarshal(payload, &logEntry)
	if err != nil {
		return fmt.Errorf("could not unmarshal request: %w", err)
	}

	f.log.Info().Str("request", logEntry.RequestID).Str("function", logEntry.Execute.FunctionID).Msg("FSM executing function")

	res, err := f.executor.ExecuteFunction(context.Background(), logEntry.RequestID, logEntry.Execute)
	if err != nil {
		return fmt.Errorf("could not execute function: %w", err)
	}

	nres := execute.NodeResult{
		Result: res,
	}

	// Execute processors.
	for _, proc := range f.processors {
		proc(logEntry, nres)
	}

	f.log.Info().Str("request", logEntry.RequestID).Msg("FSM successfully executed function")

	return res
}

func (f fsmExecutor) Snapshot() (raft.FSMSnapshot, error) {
	f.log.Info().Msg("received snapshot request")
	return nil, fmt.Errorf("TBD: not implemented")
}

func (f fsmExecutor) Restore(snapshot io.ReadCloser) error {
	f.log.Info().Msg("received snapshot restore request")
	return fmt.Errorf("TBD: not implemented")
}
