package worker

import (
	"github.com/hashicorp/go-metrics/prometheus"
)

// Tracing span names.
const (
	spanWorkOrder      = "WorkOrder"
	spanWorkOrderBatch = "WorkOrderBatch"
	spanExecute        = "Execute"
)

var (
	rollCallsSeenMetric    = []string{"node", "rollcalls", "seen"}
	rollCallsAppliedMetric = []string{"node", "rollcalls", "applied"}
	workOrderMetric        = []string{"node", "workorders"}
	workOrderBatchesMetric = []string{"node", "workorder_batches"}
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
