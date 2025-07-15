package fstore

import (
	"time"

	"github.com/hashicorp/go-metrics/prometheus"
)

const (
	defaultTimeout   = 10 * time.Second
	defaultUserAgent = "b7s"

	tracerName = "b7s.Fstore"
)

// Tracing span names.
const (
	spanInstall     = "FunctionInstall"
	spanIsInstalled = "IsFunctionInstalled"
	spanSync        = "FunctionSync"
)

var (
	functionsInstalledMetric      = []string{"fstore", "functions", "installed"}
	functionsInstalledOkMetric    = []string{"fstore", "functions", "installed", "ok"}
	functionsInstalledErrMetric   = []string{"fstore", "functions", "installed", "err"}
	functionsInstallTimeMetric    = []string{"fstore", "functions", "installation", "milliseconds"}
	functionsDownloadedSizeMetric = []string{"fstore", "functions", "installed", "size", "bytes"}
)

var Counters = []prometheus.CounterDefinition{
	{
		Name: functionsInstalledMetric,
		Help: "Number of functions installed on this node.",
	},
	{
		Name: functionsInstalledOkMetric,
		Help: "Number of successful function installs on this node in this session.",
	},
	{
		Name: functionsInstalledErrMetric,
		Help: "Number of unsuccessful functions installs on this node in this session.",
	},
	{
		Name: functionsDownloadedSizeMetric,
		Help: "Total size of (compressed) functions installed by the node in this session.",
	},
}

var Summaries = []prometheus.SummaryDefinition{
	{
		Name: functionsInstallTimeMetric,
		Help: "Total time spent downloading and installing functions",
	},
}
