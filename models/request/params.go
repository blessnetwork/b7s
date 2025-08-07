package request

const (
	// PBFT offers no resiliency towards Byzantine nodes with less than four nodes.
	pbftMinimumReplicaCount = 4
)
