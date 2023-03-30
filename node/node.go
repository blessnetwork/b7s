package node

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"

	"github.com/blocklessnetworking/b7s/host"
	"github.com/blocklessnetworking/b7s/models/blockless"
	"github.com/blocklessnetworking/b7s/node/internal/waitmap"
)

// Node is the entity that actually provides the main Blockless node functionality.
// It listens for messages coming from the wire and processes them. Depending on the
// node role, which is determined on construction, it may process messages in different ways.
// For example, upon receiving a message requesting execution of a Blockless function,
// a Worker Node will use the `Execute` component to fullfill the execution request.
// On the other hand, a Head Node will issue a roll call and eventually
// delegate the execution to the chosend Worker Node.
type Node struct {
	cfg Config

	log      zerolog.Logger
	host     *host.Host
	store    Store
	executor Executor
	function FunctionStore

	topic *pubsub.Topic
	sema  chan struct{}
	wg    *sync.WaitGroup

	rollCall         *rollCallQueue
	executeResponses *waitmap.WaitMap
}

// New creates a new Node.
func New(log zerolog.Logger, host *host.Host, store Store, peerStore PeerStore, function FunctionStore, options ...Option) (*Node, error) {

	// Initialize config.
	cfg := DefaultConfig
	for _, option := range options {
		option(&cfg)
	}

	// If we're a head node, we don't have an executor.
	if cfg.Role == blockless.HeadNode && cfg.Execute != nil {
		return nil, errors.New("head node does not support execution")
	}
	// If we're a worker node, we require an executor.
	if cfg.Role == blockless.WorkerNode && cfg.Execute == nil {
		return nil, errors.New("worker node requires an executor component")
	}

	n := Node{
		cfg: cfg,

		log:      log.With().Str("component", "node").Logger(),
		host:     host,
		store:    store,
		function: function,
		executor: cfg.Execute,

		wg:   &sync.WaitGroup{},
		sema: make(chan struct{}, cfg.Concurrency),

		rollCall:         newQueue(rollCallQueueBufferSize),
		executeResponses: waitmap.New(),
	}

	// Create a notifiee with a backing peerstore.
	cn := newConnectionNotifee(log, peerStore)
	host.Network().Notify(cn)

	return &n, nil
}

// getHandler returns the appropriate handler function for the given message.
func (n Node) getHandler(msgType string) HandlerFunc {

	switch msgType {
	case blockless.MessageHealthCheck:
		return n.processHealthCheck
	case blockless.MessageExecute:
		return n.processExecute
	case blockless.MessageExecuteResponse:
		return n.processExecuteResponse
	case blockless.MessageRollCall:
		return n.processRollCall
	case blockless.MessageRollCallResponse:
		return n.processRollCallResponse
	case blockless.MessageInstallFunction:
		return n.processInstallFunction
	case blockless.MessageInstallFunctionResponse:
		return n.processInstallFunctionResponse

	default:
		return func(_ context.Context, from peer.ID, _ []byte) error {
			return ErrUnsupportedMessage
		}
	}
}

func newRequestID() (string, error) {

	// Generate a new request/executionID.
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("could not generate new request ID: %w", err)
	}

	return uuid.String(), nil
}
