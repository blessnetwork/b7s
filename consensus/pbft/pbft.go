package pbft

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/rs/zerolog"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blocklessnetworking/b7s/host"
)

// TODO (pbft): View change.

// Replica is a single PBFT node. Both Primary and Backup nodes are all replicas.
type Replica struct {
	pbftCore
	log  zerolog.Logger
	host *host.Host

	id    peer.ID
	key   crypto.PrivKey
	peers []peer.ID

	// TODO (pbft): locking for these.

	// Keep track of seen requests. Map request to the digest.
	requests map[string]Request
	// Keep track of requests queued for execution. Could also be tracked via a single map.
	pending map[string]Request

	// Keep track of seen pre-prepare messages.
	preprepares map[messageID]PrePrepare
	// Keep track of seen prepare messages.
	prepares map[messageID]*prepareReceipts
	// Keep track of seen commit messages.
	commits map[messageID]*commitReceipts
}

// NewReplica creates a new PBFT replica.
func NewReplica(log zerolog.Logger, host *host.Host, peers []peer.ID, key crypto.PrivKey) (*Replica, error) {

	total := uint(len(peers))

	if total < MinimumReplicaCount {
		return nil, fmt.Errorf("too small cluster for a valid PBFT (have: %v, minimum: %v)", total, MinimumReplicaCount)
	}

	replica := Replica{
		pbftCore: newPbftCore(total),
		log:      log.With().Str("component", "pbft").Logger(),
		host:     host,

		id:    host.ID(),
		key:   key,
		peers: peers,

		requests:    make(map[string]Request),
		pending:     make(map[string]Request),
		preprepares: make(map[messageID]PrePrepare),
		prepares:    make(map[messageID]*prepareReceipts),
		commits:     make(map[messageID]*commitReceipts),
	}

	log.Info().Strs("replicas", peerIDList(peers)).Uint("total", total).Msg("created PBFT replica")

	// Set the message handler.
	// TODO (pbft): Split the protocols - requests should come in on the regular `b7s` protocol, while the replica communication should be on the other, `pbft cluster consensus` protocol.
	replica.setMessageHandler()

	return &replica, nil
}

func (r *Replica) setMessageHandler() {

	// We want to only accept messages from replicas in our cluster.
	// Create a map so we can perform a faster lookup.
	pm := make(map[peer.ID]struct{})
	for _, peer := range r.peers {
		pm[peer] = struct{}{}
	}

	r.host.Host.SetStreamHandler(Protocol, func(stream network.Stream) {
		defer stream.Close()

		from := stream.Conn().RemotePeer()

		// TODO (pbft): This makes sense but more locally - we have to allow requests to come in lul.
		/*
			_, known := pm[from]
			if !known {
				r.log.Info().Str("peer", from.String()).Msg("received message from a peer not in our cluster, discarding")
				return
			}
		*/

		buf := bufio.NewReader(stream)
		msg, err := buf.ReadBytes('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			stream.Reset()
			r.log.Error().Err(err).Msg("error receiving direct message")
			return
		}

		r.log.Debug().Str("peer", from.String()).Msg("received message")

		err = r.processMessage(from, msg)
		if err != nil {
			r.log.Error().Err(err).Str("peer", from.String()).Msg("message processing failed")
		}
	})
}

func (r *Replica) processMessage(from peer.ID, payload []byte) error {

	msg, err := unpackMessage(payload)
	if err != nil {
		return fmt.Errorf("could not unpack message: %w", err)
	}

	switch m := msg.(type) {

	case Request:
		return r.processRequest(from, m)

	case PrePrepare:
		return r.processPrePrepare(from, m)

	case Prepare:
		return r.processPrepare(from, m)

	case Commit:
		return r.processCommit(from, m)
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
