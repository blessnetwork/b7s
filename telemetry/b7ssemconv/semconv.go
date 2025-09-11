package b7ssemconv

import (
	"go.opentelemetry.io/otel/attribute"
)

// TODO: Add documentation for these.

const (
	ServiceRole = attribute.Key("service.role")
)

const (
	MessageType     = attribute.Key("message.type")
	MessageID       = attribute.Key("message.id")
	MessagePipeline = attribute.Key("message.pipeline")
	MessageTopic    = attribute.Key("message.topic")
	MessagePeer     = attribute.Key("message.peer")
	MessagePeers    = attribute.Key("message.peers")
)

const (
	FunctionCID       = attribute.Key("function.cid")
	FunctionMethod    = attribute.Key("function.method")
	FunctionArguments = attribute.Key("function.arguments")
)

const (
	ExecutionNodeCount = attribute.Key("execution.node.count")
	ExecutionConsensus = attribute.Key("execution.consensus")
	ExecutionRequestID = attribute.Key("execution.request.id")
)

const (
	PeerID         = attribute.Key("peer.id")
	PeerMultiaddr  = attribute.Key("peer.multiaddr")
	LocalMultiaddr = attribute.Key("peer.local.multiaddr") //TODO: Think how much useful this is.
)
