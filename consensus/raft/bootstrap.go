package raft

import (
	"errors"
	"fmt"

	"github.com/hashicorp/raft"
)

func (h *Handler) bootstrapCluster() error {

	servers := make([]raft.Server, 0, len(h.peers))
	for _, id := range h.peers {

		s := raft.Server{
			Suffrage: raft.Voter,
			ID:       raft.ServerID(id.String()),
			Address:  raft.ServerAddress(id),
		}

		servers = append(servers, s)
	}

	cfg := raft.Configuration{
		Servers: servers,
	}

	// Bootstrapping will only succeed for the first node to start it.
	// Other attempts will fail with an error that can be ignored.
	ret := h.BootstrapCluster(cfg)
	err := ret.Error()
	if err != nil && !errors.Is(err, raft.ErrCantBootstrap) {
		return fmt.Errorf("could not bootstrap cluster: %w", err)
	}

	return nil
}
