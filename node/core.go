package node

import (
	"context"

	"github.com/hashicorp/go-metrics"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/rs/zerolog"

	"github.com/blessnetwork/b7s/host"
	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/node/internal/syncmap"
	"github.com/blessnetwork/b7s/telemetry/tracing"
)

type Core interface {
	// ID returns the node ID.
	ID() string

	Logger
	Network
	Telemetry
	NodeOps
}

type Logger interface {
	Log() *zerolog.Logger
}

type Network interface {
	Host() *host.Host
	Connected(peer.ID) bool
	Messaging
}

type Messaging interface {
	Send(context.Context, peer.ID, bls.Message) error
	SendToMany(context.Context, []peer.ID, bls.Message, bool) error

	JoinTopic(string) error
	Subscribe(context.Context, string) error
	Publish(context.Context, bls.Message) error
	PublishToTopic(context.Context, string, bls.Message) error
}

type Telemetry interface {
	Tracer() *tracing.Tracer
	Metrics() *metrics.Metrics
}

type NodeOps interface {
	Run(context.Context, func(context.Context, peer.ID, string, []byte) error) error
}

type core struct {
	cfg Config

	log  zerolog.Logger
	host *host.Host

	topics *syncmap.Map[string, topicInfo]

	// Telemetry
	tracer  *tracing.Tracer
	metrics *metrics.Metrics
}

func NewCore(log zerolog.Logger, host *host.Host, opts ...Option) *core {

	cfg := DefaultConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	core := &core{
		cfg:     cfg,
		log:     log,
		host:    host,
		tracer:  tracing.NewTracer(tracerName),
		metrics: metrics.Default(),
		topics:  syncmap.New[string, topicInfo](),
	}

	return core
}

func (c *core) ID() string {
	return c.host.ID().String()
}

func (c *core) Log() *zerolog.Logger {
	return &c.log
}

func (c *core) Host() *host.Host {
	return c.host
}

func (c *core) Tracer() *tracing.Tracer {
	return c.tracer
}

func (c *core) Metrics() *metrics.Metrics {
	return c.metrics
}

func (c *core) Network() {
	c.host.Network()
}
