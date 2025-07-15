package worker

import (
	"github.com/hashicorp/go-metrics/prometheus"
)

// Tracing span names.
const (
	spanWorkOrder = "WorkOrder"
)

var (
	rollCallsSeenMetric    = []string{"node", "rollcalls", "seen"}
	rollCallsAppliedMetric = []string{"node", "rollcalls", "applied"}
	workOrderMetric        = []string{"node", "workorders"}
)

var Counters = []prometheus.CounterDefinition{
	{
		Name: rollCallsSeenMetric,
		Help: "Number of roll calls seen by the node.",
	},
	{
		Name: rollCallsAppliedMetric,
		Help: "Number of roll calls this node applied to.",
	},
	{
		Name: workOrderMetric,
		Help: "Number of work orders.",
	},
}
