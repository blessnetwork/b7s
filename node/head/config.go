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
	BatchStore:              ibs.NewBatchStore(),
}

// Config represents the Node configuration.
type Config struct {
	RollCallTimeout         time.Duration         // How long do we wait for roll call responses.
	ExecutionTimeout        time.Duration         // How long does the head node wait for worker nodes to send their execution results.
	ClusterFormationTimeout time.Duration         // How long do we wait for the nodes to form a cluster for an execution.
	DefaultConsensus        consensus.Type        // Default consensus algorithm to use.
	BatchStore              batchstore.BatchStore // Batch store for persisting batch requests
}

// BatchStore sets the batch store to be used by the head node.
func BatchStore(b batchstore.BatchStore) Option {
	return func(cfg *Config) {
		cfg.BatchStore = b
	}
}

func (c Config) Valid() error {
	return nil
}
