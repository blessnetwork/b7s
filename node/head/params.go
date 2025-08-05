package head

import (
	"time"

	"github.com/blessnetwork/b7s/consensus"
)

const (
	DefaultRollCallTimeout          = 5 * time.Second
	DefaultExecutionTimeout         = 20 * time.Second
	DefaultClusterFormationTimeout  = 10 * time.Second
	DefaultConsensusAlgorithm       = consensus.Raft
	DefaultBatchWorkItemMaxAttempts = 10
	DefaultBatchRequeueInterval     = time.Hour

	rollCallQueueBufferSize  = 1000
	executionResultCacheSize = 1000

	defaultExecutionThreshold = 0.6

	// Timeout for the context used for sending disband request to cluster nodes.
	consensusClusterSendTimeout = 10 * time.Second
)
