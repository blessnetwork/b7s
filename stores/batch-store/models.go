package batchstore

import (
	"time"
)

// TODO: Perhaps this all goes out the window and we just use the request.* types and all of that?
// TODO: Consider: ID string to UUID

type ExecuteBatchRecord struct {
	ID          string    `bson:"id,omitempty"`
	CID         string    `bson:"cid,omitempty"`
	Method      string    `bson:"method,omitempty"`
	Config      any       `bson:"config,omitempty"`
	MaxAttempts uint32    `bson:"max_attempts,omitempty"`
	Status      int32     `bson:"status"`
	CreatedAt   time.Time `bson:"created_at,omitempty"`
	UpdatedAt   time.Time `bson:"updated_at,omitempty"`
}

type ChunkRecord struct {
	ID        string    `bson:"id,omitempty"`
	BatchID   string    `bson:"batch_id,omitempty"`
	Status    int32     `bson:"status"`
	CreatedAt time.Time `bson:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty"`
}

type WorkItemRecord struct {
	ID        string    `bson:"id,omitempty"`
	BatchID   string    `bson:"batch_id,omitempty"` // TODO: Check - is it necessary? Might be good to have locality of data
	ChunkID   string    `bson:"chunk_id,omitempty"`
	Arguments []string  `bson:"arguments,omitempty"`
	Status    int32     `bson:"status"`
	Attempts  uint32    `bson:"attempts,omitempty"`
	CreatedAt time.Time `bson:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty"`
}
