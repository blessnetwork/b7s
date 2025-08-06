package host

import (
	"context"
	"fmt"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
	"golang.org/x/sync/errgroup"

	"github.com/blessnetwork/b7s/models/bls"
)

func (h *Host) ConnectToKnownPeers(ctx context.Context) error {

	err := h.ConnectToBootNodes(ctx)
	if err != nil {
		return fmt.Errorf("could not connect to bootstrap nodes: %w", err)
	}

	err = h.ConnectToDialbackPeers(ctx)
	if err != nil {
		h.log.Warn().Err(err).Msg("could not connect to dialback peers")
	}

	// Spin up a goroutine to maintain connections to boot nodes in the background.
	// In case boot nodes drops out, we want to connect back to it.
	go func(ctx context.Context) {
		ticker := time.NewTicker(h.cfg.BootNodesReachabilityCheckInterval)
		for {
			select {
			case <-ticker.C:
				err := h.ConnectToBootNodes(ctx)
				if err != nil {
					h.log.Warn().Err(err).Msg("could not connect to boot nodes")
				}

			case <-ctx.Done():
				ticker.Stop()
				h.log.Debug().Msg("stopping boot node reachability monitoring")
				return
			}
		}
	}(ctx)

	return nil
}

func (h *Host) ConnectToBootNodes(ctx context.Context) error {

	// Bootstrap nodes we try to connect to on start.
	var peers []bls.Peer
	for _, addr := range h.cfg.BootNodes {

		addrInfo, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {

			if h.cfg.MustReachBootNodes {
				return fmt.Errorf("could not get boot node address info (address: %s): %w", addr.String(), err)
			}

			h.log.Warn().Err(err).Str("address", addr.String()).Msg("could not get address info for boot node - skipping")
			continue
		}

		node := bls.Peer{
			ID:       addrInfo.ID,
			AddrInfo: *addrInfo,
		}

		peers = append(peers, node)
	}

	err := h.connectToPeers(ctx, peers)
	if err != nil {
		if h.cfg.MustReachBootNodes {
			return fmt.Errorf("could not connect to bootstrap nodes: %w", err)
		}

		h.log.Error().Err(err).Msg("could not connect to bootstrap nodes")
	}

	return nil
}

func (h *Host) ConnectToDialbackPeers(ctx context.Context) error {

	// Dial-back peers are peers we're familiar with from before.
	// We may want to limit the number of dial back peers we use.
	added := uint(0)
	addLimit := h.cfg.DialBackPeersLimit

	var peers []bls.Peer
	for _, peer := range h.cfg.DialBackPeers {

		// If the limit of dial-back peers is set and we've reached it - stop now.
		if addLimit != 0 && added >= addLimit {
			h.log.Info().Uint("limit", addLimit).Msg("reached limit for dial-back peers")
			break
		}

		// This should not happen anymore as we should have addresses, but in case it did - use the last known multiaddress.
		if len(peer.AddrInfo.Addrs) == 0 {

			ma, err := multiaddr.NewMultiaddr(peer.MultiAddr)
			if err != nil {
				h.log.Warn().Str("peer", peer.ID.String()).Str("addr", peer.MultiAddr).Msg("invalid multiaddress for dial-back peer, skipping")
				continue
			}

			h.log.Debug().Str("peer", peer.ID.String()).Msg("using last known multiaddress for dial-back peer")

			peer.AddrInfo.Addrs = []multiaddr.Multiaddr{ma}
		}

		peers = append(peers, peer)
		added++
	}

	err := h.connectToPeers(ctx, peers)
	if err != nil {
		return fmt.Errorf("could not connect to dial-back peers: %w", err)
	}

	return nil
}

func (h *Host) connectToPeers(ctx context.Context, peers []bls.Peer) error {

	// Connect to the bootstrap nodes.
	var errGroup errgroup.Group
	for _, peer := range peers {

		// Should not happen other than misconfig, but we shouldn't dial self.
		if peer.ID == h.ID() {
			continue
		}

		// Skip peers we're already connected to.
		connections := h.Network().ConnsToPeer(peer.ID)
		if len(connections) > 0 {
			continue
		}

		errGroup.Go(func() error {
			err := h.Host.Connect(ctx, peer.AddrInfo)
			// Log errors because error group Wait() method will return only the first non-nil error. We would like to be aware of all of them.
			if err != nil {
				h.log.Error().Err(err).Str("peer", peer.ID.String()).Any("addr_info", peer.AddrInfo).Msg("could not connect to bootstrap node")
				return err
			}

			h.log.Debug().Str("peer", peer.ID.String()).Any("addr_info", peer.AddrInfo).Msg("connected to peer")
			return nil
		})
	}

	// Wait until we know the outcome of all connection attempts.
	err := errGroup.Wait()
	if err != nil {
		return fmt.Errorf("some connections failed: %w", err)
	}

	return nil
}

func (h *Host) DiscoverPeers(ctx context.Context, topic string) error {

	// Initialize DHT.
	dht, err := h.initDHT(ctx)
	if err != nil {
		return fmt.Errorf("could not initialize DHT: %w", err)
	}

	discovery := routing.NewRoutingDiscovery(dht)
	util.Advertise(ctx, discovery, topic)

	h.log.Debug().Msg("host started peer discovery")

	connected := uint(0)
findPeers:
	for {
		h.log.Trace().Msg("starting peer discovery")

		// Using a list instead of a channel. If this starts getting too large switch back.
		// TODO: There's an upper limit config option, set a sane default.
		peers, err := util.FindPeers(ctx, discovery, topic)
		if err != nil {
			return fmt.Errorf("could not find peers: %w", err)
		}

		h.log.Trace().Int("count", len(peers)).Str("topic", topic).Msg("discovered peers")

		for _, peer := range peers {
			// Skip self.
			if peer.ID == h.ID() {
				continue
			}

			// Skip peers we're already connected to.
			connections := h.Network().ConnsToPeer(peer.ID)
			if len(connections) > 0 {
				continue
			}

			err = h.Connect(ctx, peer)
			if err != nil {
				h.log.Debug().Err(err).Str("peer", peer.ID.String()).Msg("could not connect to discovered peer")
				continue
			}

			h.log.Info().Str("peer", peer.ID.String()).Msg("connected to discovered peer")

			connected++

			// Stop when we have reached connection threshold.
			if connected >= h.cfg.ConnectionThreshold {
				break findPeers
			}
		}

		time.Sleep(h.cfg.DiscoveryInterval)
	}

	h.log.Info().Msg("peer discovery complete")
	return nil
}

func (h *Host) initDHT(ctx context.Context) (*dht.IpfsDHT, error) {

	// Start a DHT for use in peer discovery. Set the DHT to server mode.
	kademliaDHT, err := dht.New(ctx, h.Host, dht.Mode(dht.ModeServer))
	if err != nil {
		return nil, fmt.Errorf("could not create DHT: %w", err)
	}

	// Bootstrap the DHT.
	err = kademliaDHT.Bootstrap(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not bootstrap the DHT: %w", err)
	}

	return kademliaDHT, nil
}
