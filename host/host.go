package host

import (
	"errors"
	"fmt"
	"os"

	"github.com/asaskevich/govalidator"
	"github.com/hashicorp/go-metrics"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/rs/zerolog"
)

// Host represents a new libp2p host.
type Host struct {
	host.Host

	log zerolog.Logger
	cfg Config

	pubsub  *pubsub.PubSub
	metrics *metrics.Metrics
}

// New creates a new Host.
func New(log zerolog.Logger, address string, port uint, options ...func(*Config)) (*Host, error) {

	cfg := defaultConfig
	for _, option := range options {
		option(&cfg)
	}

	hostAddress := fmt.Sprintf("/ip4/%v/tcp/%v", address, port)
	addresses := []string{
		hostAddress,
	}

	if cfg.Websocket {

		// If the TCP and websocket port are explicitly chosen and set to the same value, one of the two listens will silently fail.
		if port == cfg.WebsocketPort && cfg.WebsocketPort != 0 {
			return nil, fmt.Errorf("TCP and websocket ports cannot be the same (TCP: %v, Websocket: %v)", port, cfg.WebsocketPort)
		}

		wsAddr := fmt.Sprintf("/ip4/%v/tcp/%v/ws", address, cfg.WebsocketPort)
		addresses = append(addresses, wsAddr)
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(addresses...),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	}

	if cfg.DisableResourceLimits {
		rcmgr, err := rcmgr.NewResourceManager(rcmgr.NewFixedLimiter(rcmgr.InfiniteLimits))
		if err != nil {
			return nil, fmt.Errorf("could not create new resource manager: %w", err)
		}

		opts = append(opts, libp2p.ResourceManager(rcmgr))
	}

	// Read private key, if provided.
	if cfg.PrivateKey != "" {
		key, err := readPrivateKey(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("could not read private key: %w", err)
		}

		opts = append(opts, libp2p.Identity(key))
	}

	if cfg.DialBackAddress != "" && cfg.DialBackPort != 0 {

		protocol, dialbackAddress, err := determineAddressProtocol(cfg.DialBackAddress)
		if err != nil {
			return nil, fmt.Errorf("could not parse dialback multiaddress (address: %s): %w", cfg.DialBackAddress, err)
		}

		externalAddr := fmt.Sprintf("/%v/%v/tcp/%v", protocol, dialbackAddress, cfg.DialBackPort)
		extAddresses := []string{
			externalAddr,
		}

		if cfg.Websocket && cfg.DialBackWebsocketPort != 0 {

			if cfg.DialBackWebsocketPort == cfg.DialBackPort {
				return nil, fmt.Errorf("TCP and websocket dialback ports cannot be the same (TCP: %v, Websocket: %v)", cfg.DialBackPort, cfg.DialBackWebsocketPort)
			}

			externalWsAddr := fmt.Sprintf("/%v/%v/tcp/%v/ws", protocol, dialbackAddress, cfg.WebsocketPort)
			extAddresses = append(extAddresses, externalWsAddr)
		}

		// Create list of multiaddrs with the external IP and port.
		var externalAddrs []ma.Multiaddr
		for _, addr := range extAddresses {
			maddr, err := ma.NewMultiaddr(addr)
			if err != nil {
				return nil, fmt.Errorf("could not create external multiaddress: %w", err)
			}

			externalAddrs = append(externalAddrs, maddr)
		}

		addrFactory := libp2p.AddrsFactory(func(addrs []ma.Multiaddr) []ma.Multiaddr {
			// Return only the external multiaddrs.
			return externalAddrs
		})

		opts = append(opts, addrFactory)
	}

	if cfg.ConnectionLimit > 0 {

		lo := cfg.ConnectionLimit
		hi := float64(cfg.ConnectionLimit) * 1.2

		cm, err := connmgr.NewConnManager(int(lo), int(hi))
		if err != nil {
			return nil, fmt.Errorf("could not create connection manager (lo: %v, hi: %v)", lo, hi)
		}

		// Protect boot nodes from pruning.
		for _, addr := range cfg.BootNodes {
			info, err := peer.AddrInfoFromP2pAddr(addr)
			if err != nil {
				continue
			}

			cm.Protect(info.ID, "boot-node")
		}

		opts = append(opts, libp2p.ConnectionManager(cm))
	}

	// Create libp2p host.
	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("could not create libp2p host: %w", err)
	}

	if cfg.EnableP2PRelay {
		log.Info().Msg("enabling p2p relay...")
		_, err = relay.New(h)
		if err != nil {
			return nil, fmt.Errorf("could not create relay: %w", err)
		}
	}

	host := Host{
		log:     log,
		cfg:     cfg,
		metrics: metrics.Default(),
	}
	host.Host = h

	return &host, nil
}

// PrivateKey returns the private key of the libp2p host.
func (h *Host) PrivateKey() crypto.PrivKey {
	return h.Peerstore().PrivKey(h.ID())
}

// PublicKey returns the public key of the libp2p host.
func (h *Host) PublicKey() crypto.PubKey {
	return h.Peerstore().PubKey(h.ID())
}

// Addresses returns the list of p2p addresses of the host.
func (h *Host) Addresses() []string {

	addrs := h.Addrs()
	out := make([]string, 0, len(addrs))

	hostID := h.ID()

	for _, addr := range addrs {
		addr := fmt.Sprintf("%s/p2p/%s", addr.String(), hostID.String())
		out = append(out, addr)
	}

	return out
}

func readPrivateKey(filepath string) (crypto.PrivKey, error) {

	payload, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	key, err := crypto.UnmarshalPrivateKey(payload)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal private key: %w", err)
	}

	return key, nil
}

// determineAddressProtocol parses the provided address and tries to determine its type. We typically expect either a IPv4, IPv6 or a hostname.
// At times it's a bit tricky to determine the address type in Go and a lot of parsers end up guessing when dealing with some more exotic variants.
func determineAddressProtocol(address string) (string, string, error) {

	if govalidator.IsIPv4(address) {
		return "ip4", address, nil
	}

	if govalidator.IsIPv6(address) {
		return "ip6", address, nil
	}

	if govalidator.IsDNSName(address) {
		return "dns", address, nil
	}

	return "", "", errors.New("could not parse address")
}
