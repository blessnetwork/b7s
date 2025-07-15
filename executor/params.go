package executor

import (
	"os"

	"github.com/hashicorp/go-metrics/prometheus"
)

const (
	defaultPermissions = os.ModePerm
	blsListEnvName     = "BLS_LIST_VARS"
	tracerName         = "b7s.Executor"
)

var (
	functionExecutionsMetric  = []string{"executor", "function", "executions"}
	functionDurationMetric    = []string{"executor", "function", "executions", "milliseconds"}
	functionCPUUserTimeMetric = []string{"executor", "function", "executions", "cpu", "user", "time", "milliseconds"}
	functionCPUSysTimeMetric  = []string{"executor", "function", "executions", "cpu", "sys", "time", "milliseconds"}
	functionOkMetric          = []string{"executor", "function", "executions", "ok"}
	functionErrMetric         = []string{"executor", "function", "executions", "err"}
)

var Counters = []prometheus.CounterDefinition{
	{
		Name: functionExecutionsMetric,
		Help: "Number of functions executed by the node.",
	},
	{
		Name: functionOkMetric,
		Help: "Number of functions successfully executed by the node.",
	},
	{
		Name: functionErrMetric,
		Help: "Number of functions executed by the node that resulted in an error.",
	},
	{
		Name: functionCPUUserTimeMetric,
		Help: "Total CPU user time this node spent executing functions in milliseconds.",
	},
	{
		Name: functionCPUSysTimeMetric,
		Help: "Total CPU sys time this node spent executing functions in milliseconds.",
	},
}

var Summaries = []prometheus.SummaryDefinition{
	{
		Name: functionDurationMetric,
		Help: "Total time this node spent executing functions - wall clock time in milliseconds.",
	},
}
