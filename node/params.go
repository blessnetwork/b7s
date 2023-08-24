package node

import (
	"errors"
	"time"

	"github.com/blocklessnetworking/b7s/consensus"
)

const (
	DefaultTopic                   = "blockless/b7s/general"
	DefaultHealthInterval          = 1 * time.Minute
	DefaultRollCallTimeout         = 5 * time.Second
	DefaultExecutionTimeout        = 10 * time.Second
	DefaultClusterFormationTimeout = 10 * time.Second
	DefaultConcurrency             = 10

	DefaultConsensusAlgorithm = consensus.Raft

	rollCallQueueBufferSize = 1000

	defaultExecutionThreshold = 0.6

	syncInterval = time.Hour // How often do we recheck function installations.
)

// Raft and consensus related parameters.
const (
	raftClusterDisbandTimeout = 5 * time.Minute
	// Timeout for the context used for sending disband request to cluster nodes.
	raftClusterSendTimeout = 10 * time.Second
)

var (
	ErrUnsupportedMessage = errors.New("unsupported message")
)
