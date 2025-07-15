package head

import (
	"github.com/hashicorp/go-metrics/prometheus"
)

// Tracing span names.
const (
	spanExecute = "Execute"
)

var (
	rollCallsPublishedMetric = []string{"node", "rollcalls", "published"}
	executionsMetric         = []string{"node", "function", "executions"}
)

var Counters = []prometheus.CounterDefinition{
	{
		Name: rollCallsPublishedMetric,
		Help: "Number of roll calls this node issued.",
	},
	{
		Name: executionsMetric,
		Help: "Number of function executions.",
	},
}
