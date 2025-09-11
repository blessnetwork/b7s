package batchstore

import (
	"time"

	"github.com/blessnetwork/b7s/models/execute"
)

type ExecuteBatchRecord struct {
	ID          string    `bson:"id,omitempty"`
	CID         string    `bson:"cid,omitempty"`
	Method      string    `bson:"method,omitempty"`
	Config      Config    `bson:"config,omitempty"`
	MaxAttempts uint32    `bson:"max_attempts,omitempty"`
	Status      int32     `bson:"status"`
	CreatedAt   time.Time `bson:"created_at,omitempty"`
	UpdatedAt   time.Time `bson:"updated_at,omitempty"`
}

// NOTE: Pulling this in as a dependency to avoid duplicate models, though I don't like the import.
type Config = execute.Config

type ChunkRecord struct {
	ID        string    `bson:"id,omitempty"`
	BatchID   string    `bson:"batch_id,omitempty"`
	Worker    string    `bson:"worker,omitempty"`
	Status    int32     `bson:"status"`
	CreatedAt time.Time `bson:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty"`
}

type WorkItemRecord struct {
	ID        string    `bson:"id,omitempty"`
	BatchID   string    `bson:"batch_id,omitempty"` // Technically not mandatory here, but is good to have locality of data.
	ChunkID   string    `bson:"chunk_id,omitempty"`
	Arguments []string  `bson:"arguments,omitempty"`
	Status    int32     `bson:"status"`
	Attempts  uint32    `bson:"attempts,omitempty"`
	Output    string    `bson:"output,omitempty"`
	CreatedAt time.Time `bson:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty"`
}

type WorkItemStatus struct {
	Output string
	Status Status
}
