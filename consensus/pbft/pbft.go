package pbft

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-metrics"
	"github.com/rs/zerolog"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"

	"github.com/blessnetwork/b7s/consensus"
	"github.com/blessnetwork/b7s/host"
	"github.com/blessnetwork/b7s/models/bls"
	"github.com/blessnetwork/b7s/telemetry/b7ssemconv"
	"github.com/blessnetwork/b7s/telemetry/tracing"
)

// TODO (pbft): Request timestamp - execution exactly once, prevent multiple/out of order executions.

// Replica is a single PBFT node. Both Primary and Backup nodes are all replicas.
type Replica struct {
	// PBFT related data.
	pbftCore
	replicaState

	cfg Config

	// Track inactivity period to trigger a view change.
	requestTimer *time.Timer

	// Components.
	log      zerolog.Logger
	host     *host.Host
	executor bls.Executor

	// Cluster identity.
	id         peer.ID
	peers      []peer.ID
	clusterID  string
	protocolID protocol.ID

	// TODO (pbft): This is used for testing ATM, remove later.
	byzantine bool

	// Telemetry
	tracer  *tracing.Tracer
	metrics *metrics.Metrics
}

// NewReplica creates a new PBFT replica.
func NewReplica(log zerolog.Logger, host *host.Host, executor bls.Executor, peers []peer.ID, clusterID string, options ...Option) (*Replica, error) {

	total := uint(len(peers))

	if total < MinimumReplicaCount {
		return nil, fmt.Errorf("too small cluster for a valid PBFT (have: %v, minimum: %v)", total, MinimumReplicaCount)
	}

	cfg := DefaultConfig
	for _, option := range options {
		option(&cfg)
	}

	replica := Replica{
		pbftCore:     newPbftCore(total),
		replicaState: newState(),

		cfg: cfg,

		log:        log.With().Str("component", "pbft").Str("cluster", clusterID).Logger(),
		host:       host,
		executor:   executor,
		clusterID:  clusterID,
		protocolID: protocol.ID(fmt.Sprintf("%s/cluster/%s", Protocol, clusterID)),

		id:    host.ID(),
		peers: peers,

		byzantine: isByzantine(),

		tracer:  tracing.NewTracer(tracerName),
		metrics: metrics.Default(),
	}

	replica.log.Info().Strs("replicas", peerIDList(peers)).Uint("n", total).Uint("f", replica.f).Bool("byzantine", replica.byzantine).Msg("created PBFT replica")

	// Set the message handlers.

	// Handling messages on the PBFT protocol.
	replica.setPBFTMessageHandler()

	return &replica, nil
}

func (r *Replica) Consensus() consensus.Type {
	return consensus.PBFT
}

func (r *Replica) Shutdown() error {
	r.host.RemoveStreamHandler(r.protocolID)
	r.stopRequestTimer()
	return nil
}

func (r *Replica) setPBFTMessageHandler() {

	// We want to only accept messages from replicas in our cluster.
	// Create a map so we can perform a faster lookup.
	pm := make(map[peer.ID]struct{})
	for _, peer := range r.peers {
		pm[peer] = struct{}{}
	}

	// Set the root span for this PBFT cluster.
	ctx := tracing.TraceContext(context.Background(), r.cfg.TraceInfo)

	r.host.Host.SetStreamHandler(r.protocolID, func(stream network.Stream) {
		defer stream.Close()

		from := stream.Conn().RemotePeer()

		// On this protocol we only allow messages from other replicas in the cluster.
		_, known := pm[from]
		if !known {
			r.log.Info().Str("peer", from.String()).Msg("received message from a peer not in our cluster, discarding")
			return
		}

		buf := bufio.NewReader(stream)
		msg, err := buf.ReadBytes('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			stream.Reset()
			r.log.Error().Err(err).Msg("error receiving direct message")
			return
		}

		r.log.Debug().Str("peer", from.String()).Msg("received message")

		err = r.processMessage(ctx, from, msg)
		if err != nil {
			r.log.Error().Err(err).Str("peer", from.String()).Msg("message processing failed")
		}
	})
}

func (r *Replica) processMessage(ctx context.Context, from peer.ID, payload []byte) (procError error) {

	// If we're acting as a byzantine replica, just don't do anything.
	// At this point we're not trying any elaborate sus behavior.
	if r.byzantine {
		return errors.New("we're a byzantine replica, ignoring received message")
	}

	// NOTE: We do double-unmarshalling here and below - here we just care about the trace info.
	ti, ok := getTraceInfoFromMessage(payload)
	if ok {
		ctx = tracing.TraceContext(ctx, ti)
	}

	msg, err := unpackMessage(payload)
	if err != nil {
		return fmt.Errorf("could not unpack message: %w", err)
	}

	ctx, span := r.tracer.Start(ctx, msgProcessSpanName(msg.Type()), trace.WithAttributes(b7ssemconv.MessagePeer.String(from.String())))
	defer span.End()
	// NOTE: This function checks the named return error value in order to set the span status accordingly.
	defer func() {
		if procError == nil {
			span.SetStatus(otelcodes.Ok, spanStatusOK)
			return
		}

		if allowErrorLeakToTelemetry {
			span.SetStatus(otelcodes.Error, procError.Error())
			return
		}

		span.SetStatus(otelcodes.Error, spanStatusErr)
	}()

	// Access to individual segments (pre-prepares, prepares, commits etc) could be managed on an individual level,
	// but it's probably not worth it. This way we just do it request by request.
	// NOTE: Perhaps lock as early as possible or force serialization. For some things we want to force in-order processing of messages,
	// e.g. `new-view` first, THEN any `preprepares` for that view.
	r.sl.Lock()
	defer r.sl.Unlock()

	err = r.isMessageAllowed(msg)
	if err != nil {
		return fmt.Errorf("message not allowed (message: %T): %w", msg, err)
	}

	switch m := msg.(type) {

	case Request:
		return r.processRequest(ctx, from, m)

	case PrePrepare:
		return r.processPrePrepare(ctx, from, m)

	case Prepare:
		return r.processPrepare(ctx, from, m)

	case Commit:
		return r.processCommit(ctx, from, m)

	case ViewChange:
		return r.processViewChange(ctx, from, m)

	case NewView:
		return r.processNewView(ctx, from, m)
	}

	return fmt.Errorf("unexpected message type (from: %s): %T", from, msg)
}

func (r *Replica) primaryReplicaID() peer.ID {
	return r.peers[r.currentPrimary()]
}

func (r *Replica) isPrimary() bool {
	return r.id == r.primaryReplicaID()
}

// helper function to to convert a slice of multiaddrs to strings.
func peerIDList(ids []peer.ID) []string {
	peerIDs := make([]string, 0, len(ids))
	for _, rp := range ids {
		peerIDs = append(peerIDs, rp.String())
	}
	return peerIDs
}

func (r *Replica) isMessageAllowed(msg interface{}) error {

	// If we're in an active view, we accept all but new-view messages.
	if r.activeView {

		switch msg.(type) {
		case NewView:
			return ErrActiveView
		default:
			return nil
		}
	}

	// We are in a view change. Only accept view-change and new-view messages.
	// PBFT also supports checkpoint messages, but we don't use those.
	switch msg.(type) {
	case ViewChange, NewView:
		return nil
	default:
		return ErrViewChange
	}
}

// cleanupState will discard old preprepares, prepares, commist and pending requests.
// Call this before updating the list of pending requests since for those we don't know
// in which view they were scheduled - we remove all of them.
func (r *Replica) cleanupState(thresholdView uint) {

	r.log.Debug().Uint("threshold_view", thresholdView).Msg("cleaning up replica state")

	// Cleanup pending requests.
	for id := range r.pending {
		delete(r.pending, id)
	}

	// Cleanup old preprepares.
	for id := range r.preprepares {
		if id.view < thresholdView {
			delete(r.preprepares, id)
		}
	}

	// Cleanup old prepares.
	for id := range r.prepares {
		if id.view < thresholdView {
			delete(r.prepares, id)
		}
	}

	// Cleanup old commits.
	for id := range r.commits {
		if id.view < thresholdView {
			delete(r.commits, id)
		}
	}
}

func isByzantine() bool {
	env := strings.ToLower(os.Getenv(EnvVarByzantine))

	switch env {
	case "y", "yes", "true", "1":
		return true
	default:
		return false
	}
}
