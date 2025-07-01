package pbft

import (
	"errors"
	"time"

	"github.com/hashicorp/go-metrics/prometheus"
	"github.com/libp2p/go-libp2p/core/protocol"
)

const (
	// Protocol to use for PBFT related communication.
	Protocol protocol.ID = "/b7s/consensus/pbft/1.0.0"

	// PBFT offers no resiliency towards Byzantine nodes with less than four nodes.
	MinimumReplicaCount = 4

	// How long do the send/broadcast operation have until we consider it failed.
	NetworkTimeout = 5 * time.Second

	// How long is the inactivity period before we trigger a view change.
	RequestTimeout = 10 * time.Second

	EnvVarByzantine = "B7S_PBFT_BYZANTINE"

	tracerName = "b7s.PBFTCluster"

	allowErrorLeakToTelemetry = false // By default we will not send processing errors to telemetry tracers.
)

var (
	ErrViewChange            = errors.New("view change in progress")
	ErrActiveView            = errors.New("replica is currently in an active view")
	ErrConflictingPreprepare = errors.New("conflicting pre-prepare")
	ErrInvalidSignature      = errors.New("invalid signature")
)

var (
	NullRequest = Request{}
)

// Tracing span status messages.
const (
	spanStatusOK  = "message processed ok"
	spanStatusErr = "error processing message"
)

var (
	pbftExecutionsTimeMetric = []string{"pbft", "execute", "milliseconds"}
)

var Summaries = []prometheus.SummaryDefinition{
	{
		Name: pbftExecutionsTimeMetric,
		Help: "Time needed to reach pBFT consensus.",
	},
}
