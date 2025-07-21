package batchstore

import (
	"time"
)

// TODO: Perhaps this all goes out the window and we just use the request.* types and all of that?
// TODO: Consider: ID string to UUID

// Description of models:
//
// ExecuteBatchRecord - single batch execution request that came into the network.
// State => in progress => completed/failed
//
// BatchRecord => 1..N ChunkRecord
// State => in progress => completed/failed
//
// ChunkRecord => 1..N WorkItemRecord
// State => in progress => completed/partial success/failed

// ChunkAssignmentRecord
// Keep track of assignments of chunks to workers

type ExecuteBatchRecord struct {
	ID     string `bson:"id,omitempty"`
	CID    string `bson:"cid,omitempty"`
	Method string `bson:"method,omitempty"`
	Config any    `bson:"config,omitempty"`

	// Values:
	// - in progress
	// - completed
	// - failed
	Status    int32     `bson:"status"`
	CreatedAt time.Time `bson:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty"`
}

type ChunkRecord struct {
	ID        string `bson:"id,omitempty"`
	RequestID string `bson:"request_id,omitempty"`
	// Values:
	// - in progress
	// - completed
	// - failed
	Status    int32     `bson:"status"`
	CreatedAt time.Time `bson:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty"`
}

type WorkItemRecord struct {
	ID        string   `bson:"id,omitempty"`
	RequestID string   `bson:"request_id,omitempty"` // TODO: Check - is it necessary? Might be good to have locality of data
	ChunkID   string   `bson:"chunk_id,omitempty"`
	Arguments []string `bson:"arguments,omitempty"`

	// Values:
	// - pending
	// - assigned
	// - completed
	// - failed
	Status    int32     `bson:"status"`
	Attempts  uint32    `bson:"attempts,omitempty"`
	CreatedAt time.Time `bson:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty"`
}

type ChunkAssignmentRecord struct {
	RequestID string    `bson:"request_id,omitempty"` // TODO: Probably not necessary
	WorkerID  string    `bson:"chunk_id,omitempty"`   // TODO: Use peer.ID
	CreatedAt time.Time `bson:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty"`
}
