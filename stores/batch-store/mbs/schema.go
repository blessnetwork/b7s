package mbs

import (
	_ "embed"
)

const (
	batchesCollection  = "b7s-batches"
	chunksCollection   = "b7s-batch-chunks"
	workItemCollection = "b7s-batch-work-items"
)

// Collections:
// - batches
// - chunks
// - work items

// TODO: Model the timestamps
// TODO: Timestamps should be mandatory

//go:embed validation/batches.json
var batchCollectionSchema []byte

//go:embed validation/chunks.json
var chunkCollectionSchema []byte

//go:embed validation/work_items.json
var workItemCollectionSchema []byte
