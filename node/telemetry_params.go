package node

import (
	"fmt"

	"github.com/hashicorp/go-metrics/prometheus"
)

// Tracing span names.
const (
	// message events
	spanMessageSend    = "MessageSend"
	spanMessagePublish = "MessagePublish"
	spanMessageProcess = "MessageProcess"

	// Tracing span status messages.
	spanStatusOK  = "message processed ok"
	spanStatusErr = "error processing message"
)

func msgProcessSpanName(msgType string) string {
	return fmt.Sprintf("%s %s", spanMessageProcess, msgType)
}

func msgSendSpanName(prefix string, msgType string) string {
	return fmt.Sprintf("%s %s", prefix, msgType)
}

var (
	messagesSentMetric      = []string{"node", "messages", "sent"}
	messagesPublishedMetric = []string{"node", "messages", "published"}
	subscriptionsMetric     = []string{"node", "topic", "subscriptions"}
	directMessagesMetric    = []string{"node", "direct", "messages"}
	topicMessagesMetric     = []string{"node", "topic", "messages"}

	messagesProcessedMetric    = []string{"node", "messages", "processed"}
	messagesProcessedOkMetric  = []string{"node", "messages", "processed", "ok"}
	messagesProcessedErrMetric = []string{"node", "messages", "processed", "err"}

	NodeInfoMetric = []string{"node", "info"}
)

var Counters = []prometheus.CounterDefinition{
	{
		Name: directMessagesMetric,
		Help: "Number of direct messages this node received.",
	},
	{
		Name: topicMessagesMetric,
		Help: "Number of topic messages this node received.",
	},
	{
		Name: subscriptionsMetric,
		Help: "Number of topics this node subscribes to.",
	},
	{
		Name: messagesSentMetric,
		Help: "Number of messages sent.",
	},
	{
		Name: messagesPublishedMetric,
		Help: "Number of messages published.",
	},
	{
		Name: messagesProcessedMetric,
		Help: "Number of messages this node processed.",
	},
	{
		Name: messagesProcessedOkMetric,
		Help: "Number of messages successfully processed by the node.",
	},
	{
		Name: messagesProcessedErrMetric,
		Help: "Number of messages processed with an error.",
	},
}

var (
	Gauges = []prometheus.GaugeDefinition{
		{
			Name: NodeInfoMetric,
			Help: "Information about the b7s node.",
		},
	}
)
