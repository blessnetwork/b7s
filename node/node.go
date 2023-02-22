package node

import (
	"context"
	"errors"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"

	"github.com/blocklessnetworking/b7s/host"
	"github.com/blocklessnetworking/b7s/models/blockless"
	"github.com/blocklessnetworking/b7s/models/response"
	"github.com/blocklessnetworking/b7s/node/internal/cache"
)

// Node is the entity that actually provides the main Blockless node functionality.
// It listens for messages coming from the wire and processes them. Depending on the
// node role, which is determined on construction, it may process messages in different ways.
// For example, upon receiving a message requesting execution of a Blockless function,
// a Worker Node will use the `Execute` component to fullfill the execution request.
// On the other hand, a Head Node will issue a roll call and eventually
// delegate the execution to the chosend Worker Node.
type Node struct {
	role      blockless.NodeRole
	topicName string

	log      zerolog.Logger
	host     *host.Host
	store    Store
	execute  Executor
	function FunctionStore
	excache  *cache.Cache

	topic *pubsub.Topic

	rollCallResponses map[string](chan response.RollCall)
	executeResponses  map[string](chan response.Execute)
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

	n := Node{
		role:      cfg.Role,
		topicName: cfg.Topic,
		excache:   cache.New(),

		log:      log,
		host:     host,
		store:    store,
		function: function,
		execute:  cfg.Execute,

		rollCallResponses: make(map[string](chan response.RollCall)),
		executeResponses:  make(map[string](chan response.Execute)),
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
			return fmt.Errorf("received an unsupported message (type: %s, from: %s)", msgType, from.String())
		}
	}
}
