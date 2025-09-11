package mocks

import (
	"context"
	"testing"

	"github.com/hashicorp/go-metrics"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/blessnetwork/b7s/host"
	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/telemetry"
	"github.com/blessnetwork/b7s/telemetry/tracing"
)

const (
	loopback = "127.0.0.1"
)

type NodeCore struct {
	IDFunc             func() string
	LogFunc            func() *zerolog.Logger
	HostFunc           func() *host.Host
	ConnectedFunc      func(peer.ID) bool
	SendFunc           func(context.Context, peer.ID, bls.Message) error
	SendToManyFunc     func(context.Context, []peer.ID, bls.Message, bool) error
	JoinTopicFunc      func(string) error
	SubscribeFunc      func(context.Context, string) error
	PublishFunc        func(context.Context, bls.Message) error
	PublishToTopicFunc func(context.Context, string, bls.Message) error
	TracerFunc         func() *tracing.Tracer
	MetricsFunc        func() *metrics.Metrics
	RunFunc            func(context.Context, func(context.Context, peer.ID, string, []byte) error) error
}

func BaselineNodeCore(t *testing.T) *NodeCore {
	t.Helper()

	tracer := tracing.NewTracer("mock-tracer")

	registry := prometheus.NewRegistry()
	sink, err := telemetry.CreateMetricSink(registry, telemetry.MetricsConfig{})
	require.NoError(t, err)

	mh, err := telemetry.CreateMetrics(sink, false)
	require.NoError(t, err)

	libp2phost, err := host.New(NoopLogger, loopback, 0)
	require.NoError(t, err)

	core := NodeCore{
		IDFunc: func() string {
			return libp2phost.ID().String()
		},
		LogFunc: func() *zerolog.Logger {
			return &NoopLogger
		},
		HostFunc: func() *host.Host {
			return libp2phost
		},
		ConnectedFunc: func(peer.ID) bool {
			return false
		},
		SendFunc: func(context.Context, peer.ID, bls.Message) error {
			return nil
		},
		SendToManyFunc: func(context.Context, []peer.ID, bls.Message, bool) error {
			return nil
		},
		JoinTopicFunc: func(string) error {
			return nil
		},
		SubscribeFunc: func(context.Context, string) error {
			return nil
		},
		PublishFunc: func(context.Context, bls.Message) error {
			return nil
		},
		PublishToTopicFunc: func(context.Context, string, bls.Message) error {
			return nil
		},
		TracerFunc: func() *tracing.Tracer {
			return tracer
		},
		MetricsFunc: func() *metrics.Metrics {
			return mh
		},
		RunFunc: func(context.Context, func(context.Context, peer.ID, string, []byte) error) error {
			return nil
		},
	}

	return &core
}

func (c NodeCore) ID() string {
	return c.IDFunc()
}

func (c NodeCore) Log() *zerolog.Logger {
	return c.LogFunc()
}

func (c NodeCore) Host() *host.Host {
	return c.HostFunc()
}

func (c NodeCore) Connected(peerID peer.ID) bool {
	return c.ConnectedFunc(peerID)
}

func (c NodeCore) Send(ctx context.Context, peerID peer.ID, msg bls.Message) error {
	return c.SendFunc(ctx, peerID, msg)
}

func (c NodeCore) SendToMany(ctx context.Context, peerIDs []peer.ID, msg bls.Message, flag bool) error {
	return c.SendToManyFunc(ctx, peerIDs, msg, flag)
}

func (c NodeCore) JoinTopic(topic string) error {
	return c.JoinTopicFunc(topic)
}

func (c NodeCore) Subscribe(ctx context.Context, topic string) error {
	return c.SubscribeFunc(ctx, topic)
}

func (c NodeCore) Publish(ctx context.Context, msg bls.Message) error {
	return c.PublishFunc(ctx, msg)
}

func (c NodeCore) PublishToTopic(ctx context.Context, topic string, msg bls.Message) error {
	return c.PublishToTopicFunc(ctx, topic, msg)
}

func (c NodeCore) Tracer() *tracing.Tracer {
	return c.TracerFunc()
}

func (c NodeCore) Metrics() *metrics.Metrics {
	return c.MetricsFunc()
}

func (c NodeCore) Run(context.Context, func(context.Context, peer.ID, string, []byte) error) error {
	return nil
}
