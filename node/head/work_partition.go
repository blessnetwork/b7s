package head

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/blessnetwork/b7s/models/request"
)

// partitionWorkBatch takes a work batch (which can contain a large number of variants of the same execution)
// and splits them among a number of workers.
//
// In the future, we may have different criteria for what gets assigned to each peer. Right now we do round robin.
func partitionWorkBatch(peers []peer.ID, req request.ExecuteBatch) {

	variants := getArgumentsVariants(req)

	// Assign arguments to a list of peers in a round robin fashion
	n := len(peers)
	a := make(map[peer.ID][][]string)
	for i, args := range variants {
		target := peers[i%n]

		a[target] = append(a[target], args)
	}

	for peer, variants := range a {
		fmt.Printf("%s\n", peer.String())
		for _, args := range variants {
			fmt.Printf("\t%s\n", strings.Join(args, " "))
		}
	}
}

func getArgumentsVariants(req request.ExecuteBatch) [][]string {

	// Create a full list of arguments - template + variants.
	variants := make([][]string, len(req.Arguments)+1)
	copy(variants[0], req.Template.Arguments)
	copy(variants[1:], req.Arguments)

	return variants
}
