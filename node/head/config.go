package head

import (
	"time"

	"github.com/blessnetwork/b7s/consensus"
	batchstore "github.com/blessnetwork/b7s/stores/batch-store"
	"github.com/blessnetwork/b7s/stores/batch-store/ibs"
)

// Option can be used to set Node configuration options.
type Option func(*Config)

// DefaultConfig represents the default settings for the node.
var DefaultConfig = Config{
	RollCallTimeout:         DefaultRollCallTimeout,
	ExecutionTimeout:        DefaultExecutionTimeout,
	ClusterFormationTimeout: DefaultClusterFormationTimeout,
	DefaultConsensus:        DefaultConsensusAlgorithm,
	WorkItemMaxAttempts:     DefaultBatchWorkItemMaxAttempts,
	RequeueInterval:         DefaultBatchRequeueInterval,
	BatchStore:              ibs.NewBatchStore(),
}

// Config represents the Node configuration.
type Config struct {
	RollCallTimeout         time.Duration    // How long do we wait for roll call responses.
	ExecutionTimeout        time.Duration    // How long does the head node wait for worker nodes to send their execution results.
	ClusterFormationTimeout time.Duration    // How long do we wait for the nodes to form a cluster for an execution.
	DefaultConsensus        consensus.Type   // Default consensus algorithm to use.
	BatchStore              batchstore.Store // Batch store for persisting batch requests
	WorkItemMaxAttempts     uint32           // How many times shoud node retry executing a work item before it gives up.
	RequeueInterval         time.Duration    // How often should head node check on batch status and requeue failed items.
}

// BatchStore sets the batch store to be used by the head node.
func BatchStore(b batchstore.Store) Option {
	return func(cfg *Config) {
		cfg.BatchStore = b
	}
}

func BatchRequeueInterval(d time.Duration) Option {
	return func(cfg *Config) {
		cfg.RequeueInterval = d
	}
}

func (c Config) Valid() error {
	return nil
}
