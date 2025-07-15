package host

import (
	"github.com/hashicorp/go-metrics/prometheus"
)

const (
	// Sentinel error for DHT.
	errNoGoodAddresses = "no good addresses"

	defaultMustReachBootNodes = false

	// Notifiee tracing span names.
	spanPeerConnected    = "PeerConnected"
	spanPeerDisconnected = "PeerDisconnected"
)

var (
	messagesSentMetric          = []string{"host", "messages", "sent"}
	messagesSentSizeMetric      = []string{"host", "messages", "sent", "bytes"}
	messagesPublishedMetric     = []string{"host", "messages", "published"}
	messagesPublishedSizeMetric = []string{"host", "messages", "published", "bytes"}
)

var Counters = []prometheus.CounterDefinition{
	{
		Name: messagesSentMetric,
		Help: "Number of messages this host sent.",
	},
	{
		Name: messagesSentSizeMetric,
		Help: "Total size of messages sent in bytes.",
	},
	{
		Name: messagesPublishedMetric,
		Help: "Number of messages this host published",
	},
	{
		Name: messagesPublishedSizeMetric,
		Help: "Total size of messages published, in bytes",
	},
}
