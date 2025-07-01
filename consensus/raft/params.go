package raft

import (
	"time"

	"github.com/hashicorp/go-metrics/prometheus"
)

// Raft and consensus related parameters.
const (
	defaultConsensusDirName = "consensus"
	defaultLogStoreName     = "logs.dat"
	defaultStableStoreName  = "stable.dat"

	defaultApplyTimeout     = 0 // No timeout.
	DefaultHeartbeatTimeout = 300 * time.Millisecond
	DefaultElectionTimeout  = 300 * time.Millisecond
	DefaultLeaderLease      = 200 * time.Millisecond

	consensusTransportTimeout = 1 * time.Minute
)

var (
	raftExecutionTimeMetric = []string{"raft", "execute", "milliseconds"}
)

var Summaries = []prometheus.SummaryDefinition{
	{
		Name: raftExecutionTimeMetric,
		Help: "Time needed to reach Raft consensus.",
	},
}
